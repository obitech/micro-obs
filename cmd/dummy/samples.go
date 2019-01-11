package main

var itemJSON = []string{`[
	{
		"name": "banana",
		"desc": "a yello fruit",
		"qty": 5
	},
	{
		"name": "water",
		"desc": "bottles of water",
		"qty": 10
	},
	{
		"name": "apple",
		"desc": "delicious",
		"qty": 15
	}
]`,
}

var orderJSON = []string{
	`{
	"items": [
		{
			"id": "BxYs9DiGaIMXuakIxX",
			"qty": 2
		},
		{
			"id": "GWkUo1hE3u7vTxR",
			"qty": 8
		}
	]
}`,
	`{
	"items": [
		{
			"id": "BxYs9DiGaIMXuakIxX",
			"qty": 4
		},
		{
			"id": "GWkUo1hE3u7vTxR",
			"qty": 5
		},
		{
			"id": "JAQU27CQrTkQCNr",
			"qty": 15
		}
	]
}`,
	`{
	"items": [
		{
			"id": "JAQU27CQrTkQCNr",
			"qty": 3
		}
	]
}`,
}
