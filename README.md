# backoff

backoff utility for cron. cron can periodically execute commands at specified intervals, but an email is sent to the cron administrator each time a command repeatedly fails. In the worst case, API calls may be banned by the web service, e.g. If you repeatedly access to web service that is under maintenance. This tool extends the interval until the next start-up according to the number of consecutive failures and suppresses the sending of annoying emails.

## Usage

```
* * * * * /path/to/the/command
```

This cron job executes the command per one minutes. next execute should be after 1 minute eventhough if the command fails.

```
* * * * * backoff /path/to/the/command
```

This cron job executes the command per one minutes. If the command fails, next execute should be after 2 minute, 4 minute, 8 minute...

## Installation

```
$ go install github.com/mattn/backoff@latest
```

## License

MIT

## Author

Yasuhiro Matsumoto (a.k.a. mattn)
