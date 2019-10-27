# Road map and comments about this work

From working on this project, here is a collection of thoughts and suggestions
for future improvements. For any questions about this, or to request help do
not hesitate to contact @damienr74 on GitHub, my email should be listed.

I think this project has the potential for the K8s community to promote best
practices. If this becomes popular, It could become easier to find
*subjectively good* configurations. This can act as a way to guide newcomers
to k8s config features that are easy to maintain, practical, and tested in some
real world environment. However, a lot of work remains to be made if this is
to happen. Extracting and ranking semantic-level information from the open
source configuration files, is definitely not trivial, and will require a lot of
though and consideration from the experts and the patterns that successful k8s
project follow. This, is outside of my scope having little to no experience with
k8s other than working on this project; however, if you have ideas I can
probably suggest approaches in order to implement it, having worked a lot on
this project.

### Improving configuration files and container configs
I did not have a lot of time to refactor the images to use configmaps for
everything. This is a good thing to improve, should be fairly easy. Another
thing that could make the user experience of launcing this could be to make all
of the go utilities be subcommands to the same binary/container image. This
would reduce the number of things that would have to be rebuilt, in order to get
it running, and it would make the application (and its components) more self
contained. (also has some disadvantages, so I'll let someone else decide.

### Adding graph metrics
From the Redis graph representation, we are able to run a multitude of graph
algorithms (not all of which are implemented).

The simplest one would be to run kruskal's algorithm to find connected
components, and to compute graph metrics on each component. Here are some of the
metrics that may be useful:

+ Average size and histograms of the sizes of each components.

+ Average size and histograms of the node with the highest in degree (rdeps) of
  each component.

+ Average size and histograms of the number of repositories in a connected
  component.

+ Any other metric that may be helpful to measure the scale of the kustomize
  import graph.

Another cool thing that may be helpful, would be to output the graph
representation of deps/rdeps. This should be fairly easy to do with graphviz/dot
so if anyone really wants this, I (damienr74) should be able to do it. Feel free
to send me an email or to @ mention me in an issue.

Note: dfs could also be used to find connected components, but I think union
find is preferable, since the results can be stored and modified very
efficiently. The only challenging part would be to implement deleting of edges
and nodes from a component efficiently, but I know it is possible to support
these operations with a union find structure.

### Implementing PageRank
The graph is set up to be able to efficiently compute PageRank since the edge
weights are real valued, and the graph representation is sparse which means that
it will fit in the memory of a single machine which will make the processing
much more efficient.

It could also be implemented as a Redis script, but I feel like there's
something fundamentally wrong with implementing PageRank in lua. :P

### Implement feature tracking
Each day, when the crawler finds and indexes these structured documents,
it should insert aggregate data to a separate index. This data could look like the
following:

```
{
  "kind": "kustomization",
  "added_identifiers": [
    {
      "identifier": "some:new:k8s:feature",
      "addedIn": [
        "docID1",
        "docID100",
        "docID45",
        ...
      ],
    },
    {
      "identifier": "another:k8s:feature",
      "documents": [
        ...
      ],
    },
    ...
  ]

  "removed_identifiers": [
    {
      "identifier": "some:deprecated:field",
      "documents": [
        ...
      ]
    }
  ]
}
```

This would make it fairly easy to get deep insight into:
- the speed at which things can effectively be deprecated.
- how many people are migrating to current best practices.
- how many documents get updated frequently/rarely.
- detailed cross sections of growth/regression over conjunctions of features.
- a world of possibilities.

This is also something that I would be interested to work on sometime soon, so
feel free to contact me (damienr74) or ask questions about this.

As needed, it could be a good idea to also aggregate past data with a larger
granularity. for instance each month, the past 30 days can be aggregated into
weekish durations, And every year these weekly aggregations can be converted
into monthly summaries depending on how much data this ends up being, and how
much you want to pay for the storage of this data.

Another cool way to compress this data would be to dynamically compress this
data into a logarithmic number of buckets with decreasing granularity. But it
seems like overkill for the amount of data that we'd likely get.

### The UI probably needs a lot of work
I'm not much of a UI/UX person and have little to no experience in developing
these types of applications. If anyone with Angular experience wants to dive in
and completely restructure the app to make the UI/UX/Code health better that
would be greatly appreciated.

### Query tuning probably still has to be adjusted
I'm also not an expert in Elasticsearch. From what I could read in the docs,
I think I've made sane decisions in converting user queries into meaningful
Elasticsearch queries, but I'm sure there are a lot of improvements that remain
to be done in order to get more accurate results.


### Some other signals that indicate the presence of a good configuration file
There are lots of heuristics that could be used to achieve this. Here are a
couple in no particular order:

+ Penalize for the number of yaml `---` document splits. I'm not sure what the
  general consensus is, but I think it's better to separate them, since it
  makes git commits less noisy, it's a trivial transformation, and it makes
  config files smaller. However, I can understand the argument that its somewhat
  practical to keep an overall view of the configurations together (maybe).

+ Penalize the number of unique identifiers in a structured document. I think
  this makes sense, since we don't want to have someone game the search engine
  to match documents with every possible path from the k8s docs. PageRank might
  help with this to some extent, but with a small corpus it would be fairly easy
  to game.

+ Assign weights to the usefulness of certain fields. It would be good to
  promote documents that use `keyRefFromConfigMap`, liveness probes, etc.

These are the main ones I can think of, but I'm sure there are a *ton* of
ways to achieve this.

If the corpus gets large enough, we might even be able to use *blockchains*,
*machine learning*, and maybe even self-driving cars.

### Add more support for indexing of other k8s/kustomize related data
One thing that jumps to mind is the use of kustomize plugins. They are easy
to track since they all have an unused global variable: `var KustomizePluggin`
it would be easy to run the pluginator command and generate godocs for each
go file with this unique identifier.

For the sake of completeness, here is the full GitHub query that we can use to
find these:
`api.github.com/search/code?q=var+KustomizePlugin+extension%3A.go&access_token=access_token`

Godoc will not show much, since most packages will be using package main, but
using pluginator we can make it a properly named package such that Godoc would
actually generate the relevant documentation.
