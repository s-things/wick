# WICK
CLI tool to make WAMP RPCs and PubSub. Useful for developing WAMP Components and their testing.

## CLI
The basic command-line interface looks like
```shell
om26er@HomePC:~$ wick
usage: wick [<flags>] <command> [<args> ...]

Flags:
  --help                     Show context-sensitive help (also try --help-long and --help-man).
  --url="ws://localhost:8080/ws"
                             WAMP URL to connect to
  --realm="realm1"           The WAMP realm to join
  --authmethod=anonymous     The authentication method to use
  --authid=AUTHID            The authid to use, if authenticating
  --authrole=AUTHROLE        The authrole to use, if authenticating
  --secret=SECRET            The secret to use in Challenge-Response Auth.
  --private-key=PRIVATE-KEY  The ed25519 private key hex for cryptosign
  --public-key=PUBLIC-KEY    The ed25519 public key hex for cryptosign
  --ticket=TICKET            The ticket when when ticket authentication
  --serializer=json          The serializer to use

Commands:
  help [<command>...]
    Show help.

  subscribe <topic>
    subscribe a topic.

  publish [<flags>] <topic> [<args>...]
    Publish to a topic.

  register <procedure> [<command>]
    Register a procedure.

  call [<flags>] <procedure> [<args>...]
    Call a procedure.
```
### Call a procedure
```shell
wick --url ws://localhost:8080/ws --realm realm1 call foo.bar
````

### Publish an event
```shell
wick --url ws://localhost:8080/ws --realm realm1 publish foo.bar arg1 arg2 --kwarg key=value --kwarg key2=value2
```

### Environment variables
Wick supports reading environment variables for all the WAMP config (realm, URL, authid, private-key...).
This is makes it effective to integrate in CI scenarios.
```shell
export WICK_URL=ws://localhost:8080/ws
export WICK_REALM=something
wick call foo.bar
```

### Supported Environment Variables
These are self-explanatory.
```shell
WICK_URL
WICK_REALM
WICK_AUTHMETHOD
WICK_AUTHID
WICK_AUTHROLE
WICK_SECRET
WICK_PUBLIC_KEY
WICK_PRIVATE_KEY
WICK_TICKET
WICK_SERIALIZER
```


## How to install
On Linux use snapd
```shell
sudo snap install wick
```
On macOS use brew
```shell
brew tap codebasepk/wick https://github.com/codebasepk/wick
brew install wick
```

## How to build
```bash
git clone git@github.com:codebasepk/wick.git
cd wick
go get github.com/codebasepk/wick/cmd/wick
go build github.com/codebasepk/wick/cmd/wick
./wick
```
