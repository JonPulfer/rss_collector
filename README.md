# rss_collector

For my submission, I have created one service that allows the addition of a feed source which
it collects on creation. Multiple sources can be added. Depending on what the ongoing purpose 
of the service, different approaches could be used to update the feed items for the feed sources.
Within the service, it could refresh the items after set interval for example.

## Admissions

I have enabled in-memory caching via the middleware in gofiber. This is using the default 1 minute expiry with
the cache key set to the full request URI (includes query args).

I haven't built this with a microservice architecture as given the time for the task I would not be able to complete
a correct solution. I have implemented distributed systems using both 2 phase commit and eventual consistency so am 
aware of the problems faced with distributed transactions and orchestration.

## 3rd party dependencies

To create the HTTP server, I have used the [gofiber](https://github.com/gofiber/fiber) framework
as it provides a lot of middleware to cater for this task. No mention was made of a particular 
authentication scheme in the task though most could be accommodated with AuthToken, JWT or a 
custom middleware incorporating the desired behaviour.

For the feed handling, I have used [gofeed](https://github.com/mmcdole/gofeed) as this simplifies
handling RSS feeds into generic data objects. If this didn't exist, it would be the approach I
would write.

## Starting

I have created a docker-compose stack of the server and a PostgreSQL database. With Docker installed
and running, from the root of the repos you can simply run: -

```shell
docker-compose up
```

To force rebuilding should code changes be made, this can be extended to: -

```shell
docker-compose up --force-recreate --build
```

When the stack has started, the server will be listening on port 8080, this can be changed 
by changing the PORT envvar in the `docker-compose.yaml` file if necessary.

I haven't specified the networking type so if there are problems connecting to the running services
from your host, that may be the cause.

## Usage

The endpoints for this service follow this convention: -

 * `GET /<objects>/` returns all <objects>
 * `GET /<objects>/:id` returns <object> with :id
 * `POST /<objects>/` where applicable creates a new object
 * `PUT /<objects>/:id` updates some detail for the object identified by :id
 * `DELETE /<objects>/:id` deletes the object identified by :id

### Adding a feed

A feed is added by making a POST request: -

```shell
curl --location --request POST 'http://localhost:8080/feeds/' \
--header 'Content-Type: application/json' \
--data-raw '{"feedUrl": "http://feeds.skynews.com/feeds/rss/technology.xml"}'
```

The response will look something like

```json
{
    "id": "c3ae3dc2-d157-4a42-9e53-c992d74486c4",
    "link": "/feeds/c3ae3dc2-d157-4a42-9e53-c992d74486c4"
}
```

### Fetching all feeds

Request: -

```shell
curl --location --request GET 'http://localhost:8080/feeds/'
```

Response: -

```json
[
    {
        "id": "8a1028dd-c9e9-490f-8748-069d8a3b0c78",
        "link": "/feeds/8a1028dd-c9e9-490f-8748-069d8a3b0c78",
        "feedUrl": "http://feeds.bbci.co.uk/news/uk/rss.xml",
        "title": "BBC News - UK",
        "categoryIDs": [
            "3e4305d5-f8d2-4a74-99d6-da875fab966c"
        ],
        "lastCollected": "2021-03-15T10:35:24.1283508Z"
    },
    {
        "id": "c3ae3dc2-d157-4a42-9e53-c992d74486c4",
        "link": "/feeds/c3ae3dc2-d157-4a42-9e53-c992d74486c4",
        "feedUrl": "http://feeds.skynews.com/feeds/rss/technology.xml",
        "title": "Tech News - Latest Technology and Gadget News | Sky News",
        "lastCollected": "2021-03-15T11:07:45.1574466Z"
    }
]
```

### Fetching a particular feed

Request: -

```shell
curl --location --request GET 'http://localhost:8080/feeds/8a1028dd-c9e9-490f-8748-069d8a3b0c78'
```

Response: - 

```json
{
    "id": "8a1028dd-c9e9-490f-8748-069d8a3b0c78",
    "link": "/feeds/8a1028dd-c9e9-490f-8748-069d8a3b0c78",
    "feedUrl": "http://feeds.bbci.co.uk/news/uk/rss.xml",
    "title": "BBC News - UK",
    "categoryIDs": [
        "3e4305d5-f8d2-4a74-99d6-da875fab966c"
    ],
    "lastCollected": "2021-03-15T10:35:24.1283508Z",
    "feedItems": [
        {
            "id": "217129b2-1a98-4cc2-8f74-6329c2441cbe",
            "sourceId": "8a1028dd-c9e9-490f-8748-069d8a3b0c78",
            "title": "Covid vaccine lowers cases in Scotland's healthcare worker families",
            "description": "Vaccination of Scotland's healthcare workers has lowered the rate of infection for people they live with.",
            "link": "https://www.bbc.co.uk/news/uk-scotland-56373252",
            "published": "2021-03-12T13:14:15Z",
            "guid": "https://www.bbc.co.uk/news/uk-scotland-56373252"
        },
        {
            "id": "8a52fc26-6f77-470b-9369-ef2e5d6f45f8",
            "sourceId": "8a1028dd-c9e9-490f-8748-069d8a3b0c78",
            "title": "Lockdown: Tourism halt 'if people from outside Wales book'",
            "description": "Self-contained accommodation can reopen in Wales from 27 March, as long as cases remain low.",
            "link": "https://www.bbc.co.uk/news/uk-wales-56375684",
            "published": "2021-03-12T15:25:49Z",
            "guid": "https://www.bbc.co.uk/news/uk-wales-56375684"
        }
    ]
}
```

### Fetching items

#### All

Request: -

```shell
curl --location --request GET 'http://localhost:8080/items'
```

Response: -

```json
[
    {
        "id": "c6dcd20a-1a8f-451c-b14d-38550118bfa4",
        "sourceId": "c3ae3dc2-d157-4a42-9e53-c992d74486c4",
        "title": "How 'religiously targeted' vaccine misinformation circulating on WhatsApp is being dealt with by faith groups",
        "description": "Faith groups are leading the fight against vaccine misinformation on what one called the \"lawless wasteland\" of WhatsApp.",
        "link": "http://news.sky.com/story/covid-19-misinformation-wars-on-whatsapp-sees-faith-groups-take-on-fake-news-12241819",
        "published": "2021-03-10T13:32:00Z",
        "guid": "http://news.sky.com/story/covid-19-misinformation-wars-on-whatsapp-sees-faith-groups-take-on-fake-news-12241819"
    },
    {
        "id": "f02b9b05-fd38-42e7-b74f-3f5c9a8a6c93",
        "sourceId": "c3ae3dc2-d157-4a42-9e53-c992d74486c4",
        "title": "Suffering a head injury before your 50s can lead to brain issues in later life, new study suggests",
        "description": "People who sustain head injuries in their 50s or younger can suffer from significant impacts to the health of their brain in later life, according to a new study led by University College London.",
        "link": "http://news.sky.com/story/suffering-a-head-injury-before-your-50s-can-lead-to-brain-issues-in-later-life-new-study-suggests-12243286",
        "published": "2021-03-11T22:46:00Z",
        "guid": "http://news.sky.com/story/suffering-a-head-injury-before-your-50s-can-lead-to-brain-issues-in-later-life-new-study-suggests-12243286"
    }
]
```

#### For a particular category

Request: -

```shell
curl --location --request GET 'http://localhost:8080/items?categoryId=3e4305d5-f8d2-4a74-99d6-da875fab966c'
```

Response: -

```json
[
    {
        "id": "56c48a22-73f2-4af0-94a0-890452460685",
        "sourceId": "8a1028dd-c9e9-490f-8748-069d8a3b0c78",
        "title": "Sarah Everard vigil: Boris Johnson 'deeply concerned' by footage",
        "description": "Police officers handcuffed women and removed them from the gathering on Clapham Common on Saturday.",
        "link": "https://www.bbc.co.uk/news/uk-56396960",
        "published": "2021-03-15T08:18:02Z",
        "guid": "https://www.bbc.co.uk/news/uk-56396960",
        "categoryIds": [
            "3e4305d5-f8d2-4a74-99d6-da875fab966c"
        ]
    },
    {
        "id": "bdf42855-313a-4b3d-a436-34e9d82e228f",
        "sourceId": "c3ae3dc2-d157-4a42-9e53-c992d74486c4",
        "title": "Facebook to label all posts about vaccines with WHO information",
        "description": "Facebook will add labels to all posts about COVID-19 vaccines to show additional information from the World Health Organisation.",
        "link": "http://news.sky.com/story/facebook-to-label-all-posts-about-vaccines-with-who-information-12246643",
        "published": "2021-03-15T09:20:00Z",
        "guid": "http://news.sky.com/story/facebook-to-label-all-posts-about-vaccines-with-who-information-12246643",
        "categoryIds": [
            "3e4305d5-f8d2-4a74-99d6-da875fab966c"
        ]
    }
]
```

#### For a particular category and source

Request: -

```shell
curl --location --request GET 'http://localhost:8080/items?categoryId=3e4305d5-f8d2-4a74-99d6-da875fab966c&sourceId=8a1028dd-c9e9-490f-8748-069d8a3b0c78'
```

Response: -

```json
[
    {
        "id": "56c48a22-73f2-4af0-94a0-890452460685",
        "sourceId": "8a1028dd-c9e9-490f-8748-069d8a3b0c78",
        "title": "Sarah Everard vigil: Boris Johnson 'deeply concerned' by footage",
        "description": "Police officers handcuffed women and removed them from the gathering on Clapham Common on Saturday.",
        "link": "https://www.bbc.co.uk/news/uk-56396960",
        "published": "2021-03-15T08:18:02Z",
        "guid": "https://www.bbc.co.uk/news/uk-56396960",
        "categoryIds": [
            "3e4305d5-f8d2-4a74-99d6-da875fab966c"
        ]
    }
]
```

### Adding a category to an item

_NB: same pattern applies for adding a category to a feed_

Request: -

```shell
curl --location --request PUT 'http://localhost:8080/items/56c48a22-73f2-4af0-94a0-890452460685' \
--header 'Content-Type: application/json' \
--data-raw '{
    "categoryIds": [
        "3e4305d5-f8d2-4a74-99d6-da875fab966c"
    ]
}'
```

Response: -

```json
{
    "id": "56c48a22-73f2-4af0-94a0-890452460685",
    "sourceId": "8a1028dd-c9e9-490f-8748-069d8a3b0c78",
    "title": "Sarah Everard vigil: Boris Johnson 'deeply concerned' by footage",
    "description": "Police officers handcuffed women and removed them from the gathering on Clapham Common on Saturday.",
    "link": "https://www.bbc.co.uk/news/uk-56396960",
    "published": "2021-03-15T08:18:02Z",
    "guid": "https://www.bbc.co.uk/news/uk-56396960",
    "categoryIds": [
        "3e4305d5-f8d2-4a74-99d6-da875fab966c"
    ]
}
```

