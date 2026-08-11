# Settings

```json-settings
{
	"tags": [
		{
			"id": "tag-5asfwgjr",
			"label": "journal",
			"color": "indigo",
			"icon": "book-open"
		},
		{
			"id": "tag-9vofk4z4",
			"label": "notebook",
			"color": "plum",
			"icon": "book-marked"
		},
		{
			"id": "tag-5wo8ixhn",
			"label": "codebook",
			"color": "teal",
			"icon": "book-text"
		},
		{
			"id": "tag-2rz0n2cv",
			"label": "framework",
			"color": "violet",
			"icon": "clipboard-list"
		},
		{
			"id": "tag-3qm4vd7p",
			"label": "analysis",
			"color": "jade",
			"icon": "bar-chart"
		}
	],
	"searches": [
		{
			"id": "search-85nzofp3",
			"title": "Uncoded highlights",
			"description": "Passages flagged with a colour and no code — the notes to myself about what the codebook does not cover yet",
			"highlight": "",
			"saved": true,
			"createdAt": 1349049600000,
			"sql": "SELECT file, id, text, reason FROM annotations WHERE code IS NULL"
		},
		{
			"id": "search-6rs6x8yi",
			"title": "Friction, journal against notebook",
			"description": "Every friction coding in date order, so the public entries and the private ones can be read side by side",
			"highlight": "",
			"saved": true,
			"createdAt": 1349136000000,
			"sql": "SELECT a.file, a.id, a.text, a.reason FROM annotations a JOIN attributes t ON t.file = a.file WHERE a.code = 'callout-5huqdng1' ORDER BY t.date"
		}
	],
	"corpusDescriptions": []
}
```
