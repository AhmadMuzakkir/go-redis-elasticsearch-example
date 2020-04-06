package elastic

const mapping = `{
	"settings":{
		"number_of_shards":2,
		"number_of_replicas":0
	},
	"mappings":{
		"properties":{
			"id":{
				"type":"keyword"
			},
			"timestamp":{
				"type":"date"
			}
		}
	}
}`
