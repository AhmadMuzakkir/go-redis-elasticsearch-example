#!/bin/bash
set -eu

docker run --rm -it --name elasticsearch -d -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" -e "network.host=_local_,_site_" -e "network.publish_host=_local_" docker.elastic.co/elasticsearch/elasticsearch:7.6.0 >/dev/null
