# Filtra

Filtra aims to extract information like lead and cycle times from Github repos for (but not limited to) projects that are doing Kanban.


# The Architecture

Filtra fetches all issues from a Github repository and provides the gathered and enriched (lead and cycle times) data as metrics. Those metrics are then stored in a PostgreSQL database and visualized with Grafana.

## Prerequisites

To get the most out of Filtra, a baseline needs to be established.

1. The project is using Github projects aka boards with multiple columns.
2. One of the columns is for planned issues. Those are not yet "in progress".
3. There might be a blocked column for issues people can't work on.
4. Lead time starts with the creation of an issue.
5. Cycle time starts when an issue was first moved out of the column for planned issues.
6. Support issues are identified with the `l3` label.
8. Bugs are identified with the `bugs` label.

Currently the planned and blocked column can be set in the [config file](https://github.com/brejoc/filtra/blob/master/config.toml).

The labels for bugs and support issues will also soon be configurable.

## Work In Progress

Github scraping is mostly done. Some metrics tweaking is still needed and additional metrics could also be gathered. Grafana is not yet automagically showing any graphs. If you know how to make this happen, please ping me or open a pull request.

If you've got ideas or want ot see additional features or metrics, head over to the [issues](https://github.com/brejoc/filtra/issues).

# Hacking on Filtra

Step 1 is only is only needed for >= Go 1.10. With Go Modules dependencies are vendorized.

1. Fetching the dependencies: `go get -d -v .`
2. Running Filtra: `go run .`
3. Access the metrics: `http://localhost:8080/metrics`

All of the metrics we are interested in start with `gh_`.


# Deployment

You can use Docker Compose to get Filtra running. But please check the [config file](https://github.com/brejoc/filtra/blob/master/config.toml) first. You should also have your Github token exported as the environment varible `$GITHUB_TOKEN`. But of course you can also paste it into the `docker-compose.yml`.

```
mkdir grafana_data
docker-compose up
```


1. Add the PostgreSQL data source.
2. Add the charts you want to see.
