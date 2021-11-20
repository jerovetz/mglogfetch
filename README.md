# mgLogFetch

mgLogFetch is a tool, which can fetch Mailgun logs, and put them to a logger service through TCP/TLS.

## Configuration

mgLogFetch uses env vars only for configuration.
- MAILGUN_API_USERNAME is the API username of mailgun (now it's _api_ by definition)
- MAILGUN_API_SECRET is the secret, which you created on Mailgun's admin
- REMOTE_LOG_HOST the host and port of the logging service where you want to push your logs to (i.e. logs.papertrailapp.com:9399) 
- OLD_THRESHOLD_SECONDS is the threshold which is used by the poller to consider log page as finished (for details see https://documentation.mailgun.com/en/latest/api-events.html#event-polling)
- MAIL_DOMAIN is your mail domain at Mailgun
- LOG_HOSTNAME is the hostname which will be put the syslog (rfc5242) formatted log
- MAILGUN_REGION the mailgun region, today is 'eu' or 'us'

## Run locally

You can use a .env or set variables in your shell.

## Building an image
```bash
 go build -o logfetcher
 docker build . // tag it if you want
```

## Contribution
All contributions are welcome.