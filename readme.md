# Artical Summarizer backend
https://artical-backend.onrender.com/

## Routes

### Register
method: POST

url: https://artical-backend.onrender.com/v1/register

post json:

```json
{
    "username": string,
    "email": string,
    "password": string,
}
```

successful response:

```json
{
    // unique id for every user
    "userID": int,
    
    // below attrs are same as request
    "username": string,
    "email":string,
}
```

### Login

method: POST

url: https://artical-backend.onrender.com/v1/login

post json:

```json
{
    "email": string,
    "password": string,
}
```

successful response:

```json
{
    // unique id for every user
    "userID": int,
    
    // below attrs are same as request
    "username": string,
    "email":string,
}
```

### Put user history
method: POST

url: https://artical-backend.onrender.com/v1/:userid/history

post json:

```json
{
    "url": string,
}
```

successful response: 

```json
{
    // give id of the inserted item
    "historyID": int
}
```

### View user history
method: GET

url: https://artical-backend.onrender.com/v1/:userid/history

successful response: 

```json
{
    // contains array of urls in LIFO order 
    "history": [
        {
            // id of history; required to delete this element
            "id": int,
            // required url to display
            "url": string,
        },
        {
            "id": int,
            "url": string,
        },
    ]
}
```

### Delete user history
method: DELETE

url: https://artical-backend.onrender.com/v1/:userid/history/:historyid

### Error

in case of error every endpoint gives error response in same structure

```json
{
    // error gives displayable error messages
    "error": string
}
```