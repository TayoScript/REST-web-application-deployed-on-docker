# PROG2005 - Assignment 2

## Renewables percentages REST API with webhooks

A simple API to retrieve data on percentages of energy that comes from renewables in different countries. Additionally, users can register webhooks to be notified when we receive a certain number of requests for a given country.

### Build & Usage

This project uses Docker Compose for deployment, please ensure you have the latest versions of Docker and Docker Compose installed.

1. First create a folder called ".credentials" inside the root folder of the project if it doesn't exist already

2. Then, copy your Firestore credentials "accountkey.json" to the folder

The project root should look like this

```
.credentials/
    |
    --- accountkey.json
cmd/
handlers/
res/
.gitignore
Dockerfile
README.md
docker-compose.yml
go.mod
go.sum
renewable-share-energy.csv
```

3. Run the following command while inside the project root directory to run the service (in the background)

```bash
docker compose up -d
```

If it does not work, double check that you have placed the credentials at the correct spot

4. Your server should now be available at port 8080. Verify by checking the status endpoint: 
```
http://<your IP>:8080/energy/v1/status/
```

### Endpoints

#### Renewables Current (/energy/v1/renewables/current/)

**Supports HTTP/REST methods**: GET  

This endpoint returns the latest available percentage of renewables for countries in our dataset. Additionally, you may filter on country and whether to include their neighbours or not.  
  
Request: /energy/v1/renewables/current/{country}?neighbours={true/false}  
  
Country can be either a 3-character code like NOR, DEU, USA etc. or a country name like Sweden, France or Canada.  
  
The neighbours parameter can be omitted, in that case it is false by default.

Example request: **/energy/v1/renewables/current/sweden?neighbours=true**

Example response:
```json
[
  {
    "name": "Finland",
    "isoCode": "FIN",
    "year": "2021",
    "percentage": 34.61129
  },
  {
    "name": "Norway",
    "isoCode": "NOR",
    "year": "2021",
    "percentage": 71.558365
  },
  {
    "name": "Sweden",
    "isoCode": "SWE",
    "year": "2021",
    "percentage": 50.924007
  }
]
```

#### Renewables History (/energy/v1/renewables/history/)

**Supports HTTP/REST methods**: GET  

This endpoint returns the historical percentages of renewables in the energy mix, including individual levels, as well as mean values for individual or selections of countries.

Path: **/energy/v1/renewables/history/**
if no {country?} is inputed all countries are returned with their mean percentages 
Example response:
```json
[
    {
        "entity": "Switzerland",
        "iso_code": "CHE",
        "percentage": 31.674854684210523
    },
    {
        "entity": "Taiwan",
        "iso_code": "TWN",
        "percentage": 3.820024382456139
    },
    {
        "entity": "United Kingdom",
        "iso_code": "GBR",
        "percentage": 3.0452300057894734
    }
]
```

{country?} refers to an optional country 3-letter code.

Example request: **/energy/v1/renewables/history/{country?}**

Example response:
```json
[
    {
        "entity": "Switzerland",
        "iso_code": "CHE",
        "percentage": 31.674854684210523
    },
    {
        "entity": "Taiwan",
        "iso_code": "TWN",
        "percentage": 3.820024382456139
    },
    {
        "entity": "United Kingdom",
        "iso_code": "GBR",
        "percentage": 3.0452300057894734
    }
```
{?begin=year&end=year} Historical percentages are returned from the year begin to year end 

Example request: **/energy/v1/renewables/history/nor?begin=1980&end=2020**

Example response:
```json
[
    {
        "entity": "Norway",
        "iso_code": "NOR",
        "year": 1980,
        "percentage": 65.252464
    },
    {
        "entity": "Norway",
        "iso_code": "NOR",
        "year": 1981,
        "percentage": 68.59788
    }
```

{?end=year} refers to year start. Historical percentages are returned from the year downwoards to the year 1965
{?begin=year}refers to year start. Historical percentages are returned from the year upwards to the year 2021 

{?sortByValue} refers to sorting percentages from lowest to highest for a specific country

Example request: **/energy/v1/renewables/history/nor?sortByValue=true**

```json
[
    {
        "entity": "Norway",
        "iso_code": "NOR",
        "year": 1970,
        "percentage": 61.510117
    },
    {
        "entity": "Norway",
        "iso_code": "NOR",
        "year": 2003,
        "percentage": 63.816036
    }
```

#### Notifications (webhooks) (/energy/v1/notifications/)

**Supports HTTP/REST methods**: GET, POST, DELETE


This endpoints handles webhooks registration. A user can register webhooks that are triggered when information about given countries is invoked, where the frequency of invocations can be set.

##### - Webhook registration
- HTTP Method: **POST**
- Path: **/energy/v1/notifications/**

**- - Request body**:
- Content Type: **application/json**

```
{			
    "url":      "(string)The URL to be triggered upon an invoked event",
    "country":  "(string)The ISO code to the country whos invocation to get notified on, if empty, i.e. "", then it applies to any country",
    "calls":    "(int)The number of invocations after which a notification is triggered, i.e. a notification is triggered for every X invocation.",
}
```
**- - - Examples:**

The user will get a notification sent to the URL given for every invocation on Iceland.
```
{
    "url": "https://webhook.site/5649d7b0-1b53-4419-912d-f4d571671bb9",
    "country": "ISL",
    "calls": 1
}
```
The user will get a notification sent to the URL given for every sixth invocation on Norway.
```
{
    "url": "https://webhook.site/5649d7b0-1b53-4419-912d-f4d571671bb9",
    "country": "NOR",
    "calls": 6
}
```
The user will get a notification sent to the URL given for every tenth invocation on any country.
```
{
    "url": "https://webhook.site/5649d7b0-1b53-4419-912d-f4d571671bb9",
    "country": "",
    "calls": 10
}
```
**- - Response**

The response will contain the registration ID of the webhook. The ID can be used to see detail information of the webhook or used to delete the webhook.

- Content Type: **application/json**
- Status code: **201**

**- - - Example:**

```
{
    "webhook_id": "btdJA6WmwWaKWBll3zNk"
}
```
##### - Deletion of webhook
- HTTP Method: **DELETE**
- Path: **/energy/v1/notification/{id}**

The **{id}** is the ID returned during registration.

**- - Response**

The response will be a message stating if deletion was succesfull or not.

- Content Type: **text/plain**
- Status Code: **200**

**- - - Example:** 

`Successfully deleted webhook`

##### - View registered webhook
- HTTP Method: **GET**
- Path: **/energy/v1/notification/{id}**

**{id}** is the ID returned during registration.

**- - Response:**
The reponse is the webhook data given during registration and the registration ID.
- Content Type: **application/json**

**- - - Example**: 

- /energy/v1/notification/6le1sdKKJmBBNvGDnzi7

```
{
    "webhook_id": "6le1sdKKJmBBNvGDnzi7",
    "url": "https://webhook.site/04ccd3e5-8f37-43e5-bcd2-6f390a1f6149",
    "country": "ISL",
    "calls": 1
}
```

##### - View all registered webhooks
- HTTP Method: **GET**
- Path: **/energy/v1/notification/**

**- - Response:**
The response is a collection of all registered webhooks.
- Content Type: **application/json**

**- - - Example:**

```
[
    {
        "webhook_id": "6le1sdKKJmBBNvGDnzi7",
        "url": "https://webhook.site/04ccd3e5-8f37-43e5-bcd2-6f390a1f6149",
        "country": "ISL",
        "calls": 1
    },
    {
        "webhook_id": "JElVOsAmyECEZWk8yKUa",
        "url": "https://webhook.site/04ccd3e5-8f37-43e5-bcd2-6f390a1f6149",
        "country": "ESP",
        "calls": 1
    },
    {
        "webhook_id": "OMOrZDS5sbqiCCwleTz0",
        "url": "https://webhook.site/eadcc7f5-d811-4ea2-ac1e-18ba85027c08",
        "country": "POL",
        "calls": 1
    },
    ...
]
```

#### Webhook invocation
When a webook is triggered upon an invocation on country it will get a notification about that.
- HTTP Method: **POST**
- Path: **URL specified by the webhook**

**- - Request Body:**
- Content Type: **application/json**


```
{			
    "webhook_id":   "(string)The ID created during registration",
    "country":      "(string)The ISO code to the invoked country.,
    "calls":        "(int)The number of invocations to the given country.",
}
```
**- - - Example:** 
```
{
  "webhook_id": "MCc9PAvDy64IESwGBRFH",
  "country": "ISL",
  "calls": 3
}
```

#### Status (/energy/v1/status/)

**Supports HTTP/REST methods**: GET  

The status interface indicates the availability of all individual services this service depends on

Path: **/energy/v1/renewables/status**

Example response:
```json

  {
    "countries_api": "200",
    "notification_db": "200",
    "webhooks": "2",
    "version": "v1",
    "uptime": 3,029
}


```

