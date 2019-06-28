# Filtra

Filtra aims to extract information like lead and cycle times from Github repos for (but not limited to) projects that are doing Kanban.


# The Architecture

In its core Filtra is a Prometheus exporter that fetches all of the issues from a Github repository and provides the gathered and enriched (lead and cycle times) data as metrics. Those metrics are then scraped by Prometheus and visualized with Grafana.

## Work In Progress

The prometheus exporter is mostly done. Only the average cycle time calculation is still missing. Prometheus knows where to fetch the Metrics, but Grafana is not yet automagically showing any graphs. If you know how to make this happen, please ping me or open a pull request.

# Hacking on Filtra

1. Fetching the dependencies: `go get -d -v .`
2. Running Filtra: `go run .`
3. Access the metrics: `http://localhost:8080/metrics`

All of the metrics we are interested in start with `gh_`.


# Deployment

You can use Docker Compose to get Filtra running. But please check the `docker-compose.yml` first. You'll have to provide the owner and repo name in the `command`. Just replace `brejoc` and `test` with somthing more relevant. And you should also have you Github token exported as the environment varible `$GITHUB_TOKEN`. But of course you can also paste it into the `docker-compose.yml`.

```
mkdir prometheus_data
mkdir grafana_data
docker-compose up
```

Prometheus needs some seconds to fetch the metrics. After that you can reach Grafana at `localhost:3000'.

1. Add the Prometheus data source.
2. Add the charts you want to see. Everything should begin with `gh_`.
