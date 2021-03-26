# ACME Client Utilities

More information: <https://github.com/hlandau/acmetool>

## Installation and Usage

You have downloaded a binary release of acmetool. Here are some simple
installation instructions:

  $ sudo cp -a bin/acmetool /usr/local/bin/

  # Run the quickstart wizard. Sets up account, cronjob, etc.
  $ sudo acmetool quickstart

  # Request the hostnames you want:
  $ sudo acmetool want example.com www.example.com

  # Now you have certificates:
  $ ls /var/lib/acme/live/example.com/

For more information on using acmetool, please see the full README at
  https://github.com/hlandau/acmetool

## Licence

    © 2015 Hugo Landau <hlandau@devever.net>    MIT License

File issues at <https://github.com/hlandau/acmetool/issues>.
