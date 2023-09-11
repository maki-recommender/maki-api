# Maki API

Core service proving REST API.


For deployment and general information refer to the [main repository](https://github.com/maki-recommender/maki).


## Configuration

This section lists the environment variables that used to configure this service. 
Variables without a default value must be assigned.

`MAKI_SqlDBConnection`

**Description:** PostgreSQL database connection url

**Example:** "host={ip} user={username} password={password} dbname=maki port={port} TimeZone=Europe/London"

`MAKI_RedisDBConnection`

**Description:** Redis database connection url

**Example** "redis://{ip}:{post}"


`MAKI_ServerAddress`

**Description:** Address where this service runs

**Default:** ":8080"


`MAKI_RecommendationServiceAddress`

**Description:** Url of the gRPC [maki model](https://github.com/maki-recommender/maki-model) server

**Default:** "0.0.0.0:50051"

`MAKI_MaxRecommendations`

**Description:** Maximum number of recommendation that a user can get in a single request (Max value of k url parameter)

**Default:** 100

`MAKI_DefaultRecommendations`

**Description:** Default number of recommendations returned in a single request (default value of k url parameter)

**Default:** 12

`MAKI_ListIsOldAfterSeconds`

**Description:** Time (in seconds) after which a user list is considered old and thus must be updated.

**Default:** 3600

`MAKI_RecommendationCacheExpireSeconds`

**Description:** Time to expire internal cache of generated recommendations. Its recommended to keep a high value to keep the load on the server as low as possible.

**Default:** 86400 (24 hours)

`MAKI_CacheClearAfterSeconds`

**Description:** Time (in seconds) after recommendation cache is completely purged for user that do not use Maki for a long time

**Default:** 604800 (7 days)


## Compiling the protocol

This section can be skipped unless some changes to the proto files need to be implemented in the application.

After installing the `requirements.dev.txt` dependencies run:

```bash
mkdir proto

python -m grpc_tools.protoc -I=./protos --python_out=./proto --pyi_out=./proto --grpc_python_out=./proto recommend_service.proto
```

