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
