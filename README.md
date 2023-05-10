# go-jwt-gin-webserver

pre requsite - you need to have go installed

1. go to the root folder and run command - go mod tidy to install an libs not there
2. then run go run main.go
3. server will get staarted at 9000 localhost
4. go to the env file and give ur mongo db. url, note the db name is given as mymongo for now, u can abstract is from env but you need to do that change
5. use the postman to call the apis for post and get data and use token as given in collection to authenticate

Learnings - 

1. gin is used as an abstraction over http std library to make it more specialized for web servers
2. the two main components to understand in how gin engine and gin context works
3. also how you can use router abstraction to group the routes and use selective middlewares on them
4. also the bson tags at struct level can be used to validate and proivide other metadata for persistence in mongodb
5. using mongo db driver in go lang and the underlying varioius bson primitive types
6. mongo db pipeline stages are used as per requirement and the projections needed
7. jwt lib used to create tokens and validate them along with user context called claims by lib
