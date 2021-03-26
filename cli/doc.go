package cli

import (
	"fmt"
	"github.com/hlandau/acmetool/storage"
	"gopkg.in/alecthomas/kingpin.v2"
)

const manPageTemplate = `{{define "FormatFlags"}}\
{{range .Flags}}\
{{if not .Hidden}}\
.TP
\fB{{if .Short}}-{{.Short|Char}}, {{end}}--{{.Name}}{{if not .IsBoolFlag}}={{.FormatPlaceHolder}}{{end}}\\fR
{{.Help}}
{{end}}\
{{end}}\
{{end}}\
{{define "FormatCommand"}}\
{{if .FlagSummary}} {{.FlagSummary}}{{end}}\
{{range .Args}} {{if not .Required}}[{{end}}<{{.Name}}{{if .Default}}*{{end}}>{{if .Value|IsCumulative}}...{{end}}{{if not .Required}}]{{end}}{{end}}\
{{end}}\
{{define "FormatCommands"}}\
{{range .FlattenedCommands}}\
{{if not .Hidden}}\
.SS
\fB{{.FullCommand}}{{template "FormatCommand" .}}\\fR
.PP
{{.Help}}
{{template "FormatFlags" .}}\
{{end}}\
{{end}}\
{{end}}\
{{define "FormatUsage"}}\
{{template "FormatCommand" .}}{{if .Commands}} <command> [<args> ...]{{end}}\\fR
{{end}}\
.TH {{.App.Name}} 8 {{.App.Version}} "acmetool"
.SH "NAME"
{{.App.Name}} - request certificates from ACME servers automatically
.SH "SYNOPSIS"
.TP
\fB{{.App.Name}}{{template "FormatUsage" .App}}
.SH "DESCRIPTION"
{{.App.Help}}
.SH "OPTIONS"
{{template "FormatFlags" .App}}\
{{if .App.Commands}}\
.SH "SUBCOMMANDS"
{{template "FormatCommands" .App}}\
{{end}}\
.SH "AUTHOR"
© 2015 {{.App.Author}} <hlandau@devever.net>  MIT License
.SH "SEE ALSO"
Documentation: <https://github.com/hlandau/acmetool>

Report bugs at: <https://github.com/hlandau/acmetool/issues>
`

var helpText = fmt.Sprintf(`acmetool is a utility for the automated retrieval, management and renewal of
certificates from ACME server such as Let's Encrypt. It emphasises automation,
idempotency and the minimisation of state.

You use acmetool by configuring targets (typically using the "want") command.
acmetool then requests certificates as necessary to satisfy the configured
targets. New certificates are requested where existing ones are soon to expire.

acmetool stores its state in a state directory. It can be specified on
invocation via the --state option; otherwise, the path in ACME_STATE_DIR is
used, or, failing that, the path "%s" (recommended).

The --xlog options control the logging. The --service options control privilege
dropping and daemonization and are applicable only to the redirector subcommand.
`, storage.RecommendedPath)

func init() {
	kingpin.CommandLine.Help = helpText
	kingpin.CommandLine.Author("Hugo Landau")
	kingpin.ManPageTemplate = manPageTemplate
}
