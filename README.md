# mongoex - Yet another way to copy mongodb data

I've looked at options like `mongomirror` (which is for whole for whole cluster, not per database/collection). I've looked at `mongodump` and `mongorestore` but there require disk space on the server that's doing the migration and is not feasible if the data is larger than the server's disk space.

So `mongoex` is a combination of both `mongodump` and `mongorestore` without the need to use any disk space and can select per database/collection if you need to.

There is only one subcomamd at the moment which is "migrate". If you dig into the source codes you will find other subcommands but these are still a WIP.

Example for copying whole databases:

```
mongoex migrate -s mongodb+srv://yourusername:yourpassword@cluster0.abc123.mongodb.net/\?retryWrites=true\&w=majority -d mongodb+srv://yourusername:yourpassword@cluster0.53yz2fy.mongodb.net/\?retryWrites=true\&w=majority --dbsrc sample_restaurants --dbdest sample_airbnb
```

Example for selective collections to copy:
```
mongoex migrate -s mongodb+srv://yourusername:yourpassword@cluster0.abc123.mongodb.net/\?retryWrites=true\&w=majority -d mongodb+srv://yourusername:yourpassword@cluster0.53yz2fy.mongodb.net/\?retryWrites=true\&w=majority --dbsrc sample_restaurants --dbdest sample_airbnb -c collection1,collection2,collection3...
```

`NOTE: -c flag for collection will get those collection from --dbsrc database`

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

Doing automated Point-in-Time Recovery to a temporary cluster
NOTE: To do this, you need to enable API Access List for your API Key - https://mongodb.com/docs/atlas/configure-api-access/#std-label-enable-api-access-list
```
mongoex atlas tempcluster pointintime --targetClusterName tempCluster --sourceProject YourSourceProject --sourceClusterName YourSourceClusterYouWantToRestoreFrom --targetProject YouTargetProject --time 1668040812
```
