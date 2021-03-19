package storageops

import (
	"crypto/x509"
	"fmt"
	"github.com/hlandau/acmetool/storage"
)

func HaveUncachedCertificates(s storage.Store) bool {
	haveUncached := false

	s.VisitCertificates(func(c *storage.Certificate) error {
		if !c.Cached {
			haveUncached = true
		}

		return nil
	})

	return haveUncached
}

// Returns the strings in ys not contained in xs.
func stringsNotIn(xs, ys []string) []string {
	m := map[string]struct{}{}
	for _, x := range xs {
		m[x] = struct{}{}
	}
	var zs []string
	for _, y := range ys {
		_, ok := m[y]
		if !ok {
			zs = append(zs, y)
		}
	}
	return zs
}

func ensureConceivablySatisfiable(t *storage.Target) {
	// We ensure that every stipulation in the satisfy section can be met by the request
	// parameters.
	excludedNames := stringsNotIn(t.Request.Names, t.Satisfy.Names)
	if len(excludedNames) > 0 {
		log.Warnf("%v can never be satisfied because names to be requested are not a superset of the names to be satisfied; adding names automatically to render target satisfiable", t)
	}

	for _, n := range excludedNames {
		t.Request.Names = append(t.Request.Names, n)
	}

	if t.Satisfy.Key.Type != "" {
		t.Request.Key.Type = t.Satisfy.Key.Type
	}
}

func DoesCertificateSatisfy(c *storage.Certificate, t *storage.Target) bool {
	if c.Revoked {
		log.Debugf("%v cannot satisfy %v because it is revoked", c, t)
		return false
	}

	if len(c.Certificates) == 0 {
		log.Debugf("%v cannot satisfy %v because it has no actual certificates", c, t)
		return false
	}

	if c.Key == nil {
		// A certificate we don't have the key for is unusable.
		log.Debugf("%v cannot satisfy %v because we do not have a key for it", c, t)
		return false
	}

	cc, err := x509.ParseCertificate(c.Certificates[0])
	if err != nil {
		log.Debugf("%v cannot satisfy %v because we cannot parse it: %v", c, t, err)
		return false
	}

	names := map[string]struct{}{}
	for _, name := range cc.DNSNames {
		names[name] = struct{}{}
	}

	for _, name := range t.Satisfy.Names {
		_, ok := names[name]
		if !ok {
			log.Debugf("%v cannot satisfy %v because required hostname %q is not listed on it: %#v", c, t, name, cc.DNSNames)
			return false
		}
	}

	if t.Satisfy.Key.Type != "" && t.Satisfy.Key.Type != c.Key.Type() {
		log.Debugf("%v cannot satisfy %v because required key type (%q) does not match (%q)", c, t, t.Satisfy.Key.Type, c.Key.Type())
		return false
	}

	log.Debugf("%v satisfies %v", c, t)
	return true
}

func FindBestCertificateSatisfyingFromCache(certArray []*storage.Certificate, t *storage.Target) (*storage.Certificate, error) {
	var bestCert *storage.Certificate

	for _, c := range certArray {
		if DoesCertificateSatisfy(c, t) {
			isBetterThan, err := CertificateBetterThan(c, bestCert)
			if err != nil {
				continue
			}

			if isBetterThan {
				log.Tracef("findBestCertificateSatisfying: %v > %v", c, bestCert)
				bestCert = c
			} else {
				log.Tracef("findBestCertificateSatisfying: %v <= %v", c, bestCert)
			}
		}
	}

	if bestCert == nil {
		return nil, fmt.Errorf("%v: no certificate satisfies this target", t)
	}

	return bestCert, nil
}

func FindBestCertificateSatisfying(s storage.Store, t *storage.Target) (*storage.Certificate, error) {
	var bestCert *storage.Certificate

	err := s.VisitCertificates(func(c *storage.Certificate) error {
		if DoesCertificateSatisfy(c, t) {
			isBetterThan, err := CertificateBetterThan(c, bestCert)
			if err != nil {
				return err
			}

			if isBetterThan {
				log.Tracef("findBestCertificateSatisfying: %v > %v", c, bestCert)
				bestCert = c
			} else {
				log.Tracef("findBestCertificateSatisfying: %v <= %v", c, bestCert)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if bestCert == nil {
		return nil, fmt.Errorf("%v: no certificate satisfies this target", t)
	}

	return bestCert, nil
}

func CertificateBetterThan(a, b *storage.Certificate) (bool, error) {
	if b == nil || a == nil {
		return (b == nil && a != nil), nil
	}

	if len(a.Certificates) == 0 || len(b.Certificates) == 0 {
		return false, fmt.Errorf("need two certificates to compare")
	}

	ac, err := x509.ParseCertificate(a.Certificates[0])
	bc, err2 := x509.ParseCertificate(b.Certificates[0])
	if err != nil || err2 != nil {
		if err == nil && err2 != nil {
			log.Tracef("certBetterThan: parseable certificate is better than unparseable certificate")
			return true, nil
		}

		return false, nil
	}

	isAfter := ac.NotAfter.After(bc.NotAfter)
	log.Tracef("certBetterThan: (%v > %v)=%v", ac.NotAfter, bc.NotAfter, isAfter)
	return isAfter, nil
}

func CertificateNeedsRenewing(c *storage.Certificate, t *storage.Target) bool {
	if len(c.Certificates) == 0 {
		log.Debugf("%v: not renewing because it has no actual certificates (???)", c)
		return false
	}

	cc, err := x509.ParseCertificate(c.Certificates[0])
	if err != nil {
		log.Debugf("%v: not renewing because its end certificate is unparseable", c)
		return false
	}

	renewTime := renewTime(cc.NotBefore, cc.NotAfter, t)
	needsRenewing := !InternalClock.Now().Before(renewTime)

	log.Debugf("%v needsRenewing=%v notAfter=%v", c, needsRenewing, cc.NotAfter)
	return needsRenewing
}

// This is used to detertmine whether to cull certificates.
func CertificateGenerallyValid(c *storage.Certificate) bool {
	// This function is very conservative because if we return false
	// the certificate will get deleted. Revocation and expiry are
	// good reasons to delete. We already know the certificate is
	// unreferenced.

	if c.Revoked {
		log.Debugf("%v not generally valid because it is revoked", c)
		return false
	}

	if len(c.Certificates) == 0 {
		// If we have no actual certificates, give the benefit of the doubt.
		// Maybe the certificate is undownloaded.
		log.Debugf("%v has no actual certificates, assuming valid", c)
		return true
	}

	cc, err := x509.ParseCertificate(c.Certificates[0])
	if err != nil {
		log.Debugf("%v cannot be parsed, assuming valid", c)
		return false
	}

	if !InternalClock.Now().Before(cc.NotAfter) {
		log.Debugf("%v not generally valid because it is expired", c)
		return false
	}

	return true
}
