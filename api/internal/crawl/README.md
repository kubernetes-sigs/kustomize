## What is this?
### In short
Be the GoDoc.org of k8s configuration files.

### More explicitly
Support k8s document indexing from open-source configurations in order to make
it easy for people to learn to use a new feature, explore k8s configs in a
central hub, and see some metrics about kustomize use.

We want people to be able to support three main classes of queries:

1. Structured document queries: how should I use the following fields
   - Grace periods: `spec:template:spec:terminationGracePeriod`?
   - Kustomize inline patch: `patches:patch`?

2. Key value queries: how should I use this more specific use case of a
   structure configuration.
   - HorizontalPodAutoScalers: `kind=HorizontalPodAutoScaler`?
   - Patches on StatefulSets: `patches:target:kind=StatefulSet`?

3. Full text search: search the comments and the document text from any
   type of k8s config file.

## Road map
There is a lot that can be added in order to improve the state of this
application. Some more details along with general thoughts and comments can be
found in the Roadmap.md file in this directory. This README contains only
what can be considered as mostly complete and iterable parts of this project.

## Running this project
Everything is configured using kubernetes, so it should be easy for people to
spin this up on any k8s cluster. Everything should just work (TM).

The config files live in the `config` directory.

```
config
├── base
│   └── kustomization.yaml
├── crawler
│   ├── base
│   │   ├── github_api_secret.txt
│   │   └── kustomization.yaml
│   ├── cronjob
│   │   ├── cronjob.yaml
│   │   └── kustomization.yaml
│   └── job
│       ├── job.yaml
│       └── kustomization.yaml
├── elastic
│   └── ...
├── redis
│   ├── document_keystore
│   │   ├── kustomization.yaml
│   │   ├── redis.yaml
│   │   └── service.yaml
│   └── http_cache
│       ├── kustomization.yaml
│       ├── redis.yaml
│       └── service.yaml
├── webapp
│   ├── backend
│   │   ├── deployment.yaml
│   │   ├── kustomization.yaml
│   │   └── service.yaml
│   └── frontend
│       ├── deployment.yaml
│       ├── kustomization.yaml
│       └── service.yaml
└── schema_files
    └── kustomization_index
        ├── es_index_mappings.json
        └── es_index_settings.json
```

To get everything up and running you have to:

1. Get some instance of elasticsearch working... and configure the
   configmapGenerator in `config/base` to point to the right endpoint(s). The
   configurations that need this value to be populated are the following:
    - `config/crawler/cronjob` to run periodic crawls.
    - `config/crawler/job` to run crawls on demand.
    - `config/webapp/backend` to run the search server.

2. Configure the elasticsearch indices:
```
kustomize build config/schema_files/kustomization_index | kubectl apply -f -
```
This will run a `curl` command that reads json data from a ConfigMap. This will
setup the schema.  If you want to make more complex modifications to the
schema, you should refer to the elastic docs to figure out whether the mapping
can be added to the current index, or whether you will need to copy the
existing index into a different one with the appropriate mappings. Modifications
can be made by using the elasticsearch go library and writing a simple program,
or it can be made with any http command to the appropriate server endpoint from
within the cluster. Unfortunately I did not have the time to write a few helper
tools for this. Feel free to contact me if you need help with modifying
elasticsearch configs, I'm by no means an expert, but I can try to help.

3. (Optional) run the redis http chache for the crawler:
```
kubectl apply -k config/redis/http_cache
```
  This will create a deployment for the cache, and a service. The crawler should
  be configured to connect to the `http_cache` if it exists, but you can always
  check the logs to make sure it connects, and that the identifiers match in the
  crawler configuration and for the service endpoint.

  The please be aware that the cache does not have a persistent volume.

4. Configure the main redis instance:
```
kubectl apply -k config/redis/document_keystore
```
  This will create a StatefulSet with a volume of 4GiB for a redis instance.

5. Get an access token from GitHub.

To be able to kindly ask GitHub for it's data on k8s config files, you'll need
to create an access\_token. From my understanding, this is the only way to do
these code search queries (without first specifying a repository).

To generate a token, go to your GitHub's account in Settings > Developer
Settings > Personal access tokens. It should look like this.

![GitHub Token 1](
https://sigs.k8s.io/kustomize/internal/tools/pictures/github_token.png)

From here you want to generate a new token and  have the following
configuration:

![GitHub Token 1](
https://sigs.k8s.io/kustomize/internal/tools/pictures/token_config.png)

If you have uses for any other data from this token, (org data, or something
else) you can pick and choose, but be careful since it can grant this
application access to your notifications, etc. However, any such extension
is explicitly a non-goal and would not be maintained by this project.

6. Launch the crawler:
```
kustomize build config/crawler/cronjob | kubectl apply -f -
```
This will periodically run the crawler every day according to the cron timing
rules in the cronjob.yaml file.

Instead, to get the crawler running now, you can run:
```
kustomize build config/crawler/cronjob | kubectl apply -f -
```
which will launch a non-periodic version of the crawler. It will take a few
minutes for the crawler to split the search, but then config files should
start to get populated within 20 minutes. It may take a while to do the
first crawl, since it has to fetch rate-limited endpoints for each new file it
finds. It should get significantly faster to update in the future.

5. Launch the search backend
```
kustomize build config/webapp/backend | kubectl apply -f -
```

6. Launch the search frontend
```
kustomize build config/webapp/frontend | kubectl apply -f -
```

## Notes about the components

### Elasticsearch
I will add a basic working setup soon. I just did the lazy thing and used an
already packaged solution. Most clouds will provide their own elastic
environments, however, Elasticsearch is also working on their own
implementation of a
![k8s operator](https://www.elastic.co/elasticsearch-kubernetes), which might
be worth checking out. Please note that it comes with its own license
agreement.

### Redis
There are two Redis instances that are used in this application.

One of them is configured to have on disk persistence, so make sure to have
that set up in your kubernetes cluster. Also note that it is running on a
single master node (i.e. it does not automatically shard keys to multiple head
nodes as part of a highly available cluster). Since it's storing a sparse
graph, I can't imagine this being much of an issue, but it's probably worth
mentioning.

The other Redis instance is running as a HTTP (RFC 7234) cache for etags from
GitHub (or any other document store from which we could crawl/index). This one
does not require full persistent storage on disk. The caching strategy is an
LRU cache which is probably a good starting point. It might be worth it to
investigate other cache policies, but I think LRU will work well since
documents may or may not expire anyway, and the amount of memory allocated for
keys is fairly large, so eviction of frequently used documents seems unlikely
anyway.

### Nginx + Angular
There is a Dockerfile included for generating the container image with Nginx
(using the default package) and adding all of the supporting compiled angular
files. Any modifications to the code-base should be compatible with this setup,
so all that's needed is to rebuild the container image, and possibly modify
the image tags in the k8s file.

### Supporting Go binaries
There are a few go binaries that each have their own Dockerfile to build
containers in which to run them on k8s, namely the crawler and the search
service. Their configurations are not optimal (read: needs to be cleaned up),
but they are functional.

## Technical details

### Overall design and imlpementation

There are a few components that are all running together in order to get
the overall application to work smoothly. This section will provide a brief
overview of each component with the following sections going into more details.

The overall structure is outlined in the following figure:
![overview](
https://sigs.k8s.io/kustomize/internal/tools/pictures/sys_arch.png)

#### Crawler
The leftmost component consists of a crawler with an http cache of GitHub
queries does two things, it first looks at the list of documents in
elasticsearch and tries to update them. In doing so, it maintains a set of
newly updated files to exclude them from other parts of the crawl.

To find newly added documents, the crawler crawls any new dependencies
introduced in the document updating step and it also queries GitHub for the
most recently indexed kustomization.\* files. Each new file will be processed
for efficient text queries and put into the document index. Any new dependency
will also incur more crawl operations. Finally, a graphical
representation of the documents and their dependencies is built in Redis to be
used for graph algorithms such as PageRank and component analysis.

#### Data library
There are a few helper libaries for dealing with Elasticsearch, Redis and
documents. This is not persistent, nor is it centralized. They act as small
components that help to package common pieces of code. Eventually it may make
sense to merge all of it together and make a proper persistent model around
this while providing an external API for document insertion/deletion. But
that is definitely out of scope in terms of getting this to run. However
there are limitations with the current model in terms of minimizing the
API surface for the different components of the application. For now this
problem is mostly mitigated by having the query server only connected to
a data node of the Elasticsearch cluster, but the problem of knowing what
is accessible and what isn't is left to the programmer instead of being
clearly and explicitly supported by the API.

#### Server
Uses the data library to communicate with the data store and answer queries.
Processes the user entered text queries into somewhat optimized elasticsearch
queries. Provides a few endpoints to get different metrics and to eventually
allow for registration of remote repositories.

This application has an exposing service in order to allow users of the
application access to queries and the results.

#### Nginx + Angular
Communicates directly with the backend server to forward user queries and
their results. Presents the results on an interface. It's still pretty simple
looking but it seems usable (to me).


### Crawling GitHub
With the use of API keys, GitHub allows account owners to search for files
using their API.

The search endpoints allow for the use of metadata search
that is fairly useful/powerful. For instance they provide a `filename:` keyword
that permits us to look for `kustomization.yaml`, `kustomization.yml`, etc.
This enables the fetching of a list of kustomization documents, from which
we can get the actual content from another endpoint
(raw.githubusercontent.com).

However, the search API is fairly limited. There is a restriction to the number
of documents that can be retrieved from this method. One possible way to
mitigate this would be to periodically query GitHub for results, sorted by the
last indexed time. This would allow you to collect most documents from this
point forwards. The downside to this is that it may require a large number of
requests to their API since you cannot know when new files will be added.
Furthermore, there is a possibility that you would not be able to get all of
files either, depending on the velocity of growth.

The approach that was taken to mitigate this is to use the `filesize:` keyword
and to shard the search space into contiguous buckets of appropriate size in
order to get all of the documents. This is fairly efficient, since you can find
a good enough way to shard the documents in
`lg(max file size) * number of documents / 1000` API queries. Moreover, since
queries are paginated with at most 100 results per query, this solution is
competitive with getting the optimal (non-contiguous) sharding of result sets.
Furthermore, filesize queries can be cached to minimize the total number of
queries called to the API in order to shard the search space. This is done by
querying for file size intervals that always start with 0..X and binary
searching over the `filesize:` space. This will allow you to reuse a lot of
queries when you're looking for the next range, since it is upper bounded and
lower bounded to a smaller number of queries within a range that has also been
queried. I think this is only true because filesizes are power law distributed,
so searches will typically require less queries as they progress from left to
right.

However, this method in no way depends on intervals of the form 0..X, as
the number of documents in the many intervals of the range search could be
added together to also make this work. This approach just seemed simpler to
implement, maintain, and debug so it was preferred.

To get an idea of how efficient this method is, to shard the search space of
7000 documents, it will only take ~90 API range queries which should only take
a few minutes. While actually fetching the documents and their relevant
metadata (creation time, etc.) will take several hours. Furthermore, this
could be made more efficient if a prior distribution is approximated.
This prior could be scaled to the number of documents that need to be fetched,
and then finding a shard that has an adequate number of requests, will only
take a few queries per shard. It could probably be supported in a constant
number of size queries if the size of each shard is halved which shouldn't
have terrible performance impact for the retrieval. However, there where
more pressing things to implement. I might revisit this later.

### Document Indexing and Processing
In order to support simple text queries the structured documents must be
processed in some way that makes searching them easy. The current method
is to recursively traverse the map of configurations to generate each sub-path
and each key-value pair for the leaf nodes of the recursion tree.

However, note that this means that a document has to be valid yaml/json
format in order for indexing to happen. The rest of the document is treated
as mostly text and uses default text settings from Elasticsearch.

What this means is that for the following yaml document:

```yaml
resources:
- service.yaml
- deployment.yaml

configmapGenerator:
- name: app-configuration
  files:
  - config.yaml

patchesJson6902:
- target:
    version: v1
    kind: StatefulSet
    name: ss-name
  path: ss-patch.yaml
- target:
    version: v1
    kind: Deployment
    name: dep-name
  path: dep-patch.yaml
```

the following flattened structure would look like:
```
{
  "identifiers": [
    "resources",
    "configmapGenerator",
    "configmapGenerator:name",
    "configmapGenerator:files",
    "patchesJson6902",
    "patchesJson6902:target",
    "patchesJson6902:target:version",
    "patchesJson6902:target:kind",
    "patchesJson6902:target:name",
    "patchesJson6902:path",
  ],
  "values": [
    "resources=service.yaml",
    "resources=deployment.yaml",
    "configmapGenerator:name=app-configuration",
    "configmapGenerator:files=config.yaml",
    "patchesJson6902:target:version=v1",
    "patchesJson6902:target:kind=StatefulSet",
    "patchesJson6902:target:name=ss-name",
    "patchesJson6902:path=ss-patch.yaml",
    "patchesJson6902:target:kind=Deployment",
    "patchesJson6902:target:name=dep-name",
    "patchesJson6902:path=dep-patch.yaml",
  ],
  ...
}
```

Note that unique paths and values are deduplicated.

On the search side, exact queries will be prioritized, but the document paths
and key=value pairs will also be analyzed with 3-grams to have some amount of
fuzzy search. The reason that a Levenshtein-Distance was not used instead, is due
to searching multiple fields at the same time, which is a use case where
Elasticsearch does not support proper fuzzy searching.

### Document Search
Given a text query, each token is considered separately. Each token will be fed
through a handful of analyzers on the Elasticsearch side, and will be compared
with the reverse document index of each document fields. It will then determine
the best matching documents. Text ordering is largely insignificant. This makes
sense for the structured search, but may leave room for improvement for the
text only search within the document.

Each token _must_ be matched, so each white space character acts as a
conjunction of individual queries. There are also ways of telling
Elasticsearch that some things _should_ match, but I think for now it makes
more sense to leave it as is.

I think this behavior is sufficient to make the search feel fairly intuitive
while providing support for fairly complex use cases.

### Metrics Computation
From the each kustomization document that is indexed, we can find it's
resources that are publicly available. This includes other kustomizations.
From this, we can build a directed graph of dependencies and reverse
dependencies.

This opens up the possibility to add a plethora of graph metrics that can
give the project maintainers feedback and insight into how people are using
their tools.

Some of these are useful such as getting an idea for how large the dependency
graphs actually grow in practice, and can be used to find _popular_
kustomizations within the corpus. This lends itself to implementing PageRank
to help bubble up popular results as good search results. I unfortunately
did not have the time to implement the algorithm, but I do plan to revisit
this sometime soon to add a few good and efficient implementations of useful
graph algorithms that would be useful to have. See the Roadmap.md for a more
complete list of features that could be added and how I think they could be
implemented.
