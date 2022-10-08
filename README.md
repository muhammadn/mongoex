# mongoex - Yet another way to copy mongodb data

There is only one subcommit at the moment and this tool is heavily in development.

Subcommand for now is "migrate"

Example for copying whole databases:

```
mongoex migrate -s mongodb+srv://yourusername:yourpassword@cluster0.abc123.mongodb.net/\?retryWrites=true\&w=majority -d mongodb+srv://yourusername:yourpassword@cluster0.53yz2fy.mongodb.net/\?retryWrites=true\&w=majority --dbsrc sample_restaurants --dbdest sample_airbnb
```

Example for selective collections to copy:
```
mongoex migrate -s mongodb+srv://yourusername:yourpassword@cluster0.abc123.mongodb.net/\?retryWrites=true\&w=majority -d mongodb+srv://yourusername:yourpassword@cluster0.53yz2fy.mongodb.net/\?retryWrites=true\&w=majority --dbsrc sample_restaurants --dbdest sample_airbnb -c collection1,collection2,collection3...
```
Try using `--help` flag to see the options, normally you will see this:

```
mongoex is a tool to migrate mongodb data in real time.
This tool helps to quickly do migrations to move data, especially from production to pre-prod for testing

Usage:
  mongoex [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  migrate     Migrates data from a source MongoDB to destination MongoDB

Flags:
  -h, --help   help for mongoex

Use "mongoex [command] --help" for more information about a command.
```
