# Findings — Cape Aldis, March to October 1912

Eight journal files, one private notebook, seven codes, and a hundred and eleven codings. What follows is what the counting shows and, in one case, what it does not license.

Every figure below draws from the journals only. The notebook covers the same eight months in a single file, so it carries one date and would drop its whole season into whichever month that date fell in. It is read against the journals further down instead.

## The season

```json-chart
{
	"id": "chart-4edlfn6v",
	"caption": {
		"label": "Codings per month by code, the eight journal files"
	},
	"query": "SELECT strftime(t.date, '%Y-%m') AS month, a.code AS code, count(*) AS count FROM annotations a JOIN attributes t ON t.file = a.file WHERE a.code IS NOT NULL AND t.type = 'journal' GROUP BY 1, 2 ORDER BY 1",
	"spec": {
		"type": "stacked-bar",
		"x": {
			"field": "month",
			"label": "Month"
		},
		"y": {
			"field": "count",
			"label": "Codings"
		},
		"series": "code",
		"color": "{code:color}",
		"tooltip": "**{code}**\\n{count} codings in {month}",
		"bands": [
			{
				"from": "1912-04",
				"to": "1912-08",
				"label": "Polar night"
			}
		]
	}
}
```

The shaded band is the polar night. Its edges are April and August entire, because a band snaps to whole months on this axis and the sun in fact went on 26 April and came back on 20 August — so both edge months are partly lit and the band overstates the dark at each end by a few days.

Volume rises after the light returns rather than during the dark. September carries twice what March does. The obvious reading is that more happened once the party could move, and that is probably right, but it is worth saying that the journal is also longer in September, so some of this is a corpus that grew rather than a winter that intensified.

## The same wind, twice

Weather is the subject the corpus never leaves, and two codes divide it. One is weather written as something with intent; the other is weather written as a number. The words are the same words.

```json-chart
{
	"id": "chart-0wpvpsm6",
	"caption": {
		"label": "The same weather, coded two ways"
	},
	"query": "SELECT strftime(t.date, '%Y-%m') AS month, a.code AS code, count(*) AS count FROM annotations a JOIN attributes t ON t.file = a.file WHERE t.type = 'journal' AND a.code IN ('callout-4tfndxpe', 'callout-1zk5b338') GROUP BY 1, 2 ORDER BY 1",
	"spec": {
		"type": "line",
		"x": {
			"field": "month",
			"label": "Month"
		},
		"y": {
			"field": "count",
			"label": "Codings"
		},
		"series": "code",
		"color": "{code:color}",
		"tooltip": "**{code}**\\n{count} codings in {month}",
		"bands": [
			{
				"from": "1912-04",
				"to": "1912-08",
				"label": "Polar night"
			}
		]
	}
}
```

The lines cross in August, in the month the sun came back. Through the dark, the weather is the only thing happening and it acquires a character: it aims, it leans, it would rather not be freed. Once the party can act on it, it goes back to being a figure in a log.

No keyword search separates these two. Barrow's entry of 5 June carries both, four lines apart, in one hand.

## Ritual against friction

```json-chart
{
	"id": "chart-1hk1dwrs",
	"caption": {
		"label": "Ritual maintenance against friction displaced"
	},
	"query": "SELECT strftime(t.date, '%Y-%m') AS month, a.code AS code, count(*) AS count FROM annotations a JOIN attributes t ON t.file = a.file WHERE t.type = 'journal' AND a.code IN ('callout-6rcq1wsu', 'callout-5huqdng1') GROUP BY 1, 2 ORDER BY 1",
	"spec": {
		"type": "line",
		"x": {
			"field": "month",
			"label": "Month"
		},
		"y": {
			"field": "count",
			"label": "Codings"
		},
		"series": "code",
		"color": "{code:color}",
		"tooltip": "**{code}**\\n{count} codings in {month}",
		"bands": [
			{
				"from": "1912-04",
				"to": "1912-08",
				"label": "Polar night"
			}
		]
	}
}
```

Ritual maintenance peaks at midwinter and friction displaced is at its lowest in the same month. The reading that suggests itself is that ceremony suppresses conflict.

That reading is not available from this figure. Midwinter dinner is fixed by the calendar and cannot be a response to anything, so the anti-phase cannot be cause in that direction, and the corpus offers a plainer explanation: preparing it occupied the fortnight either side, and men who are conspiring about a menu are not quarrelling about a damper. The figure is real. The story is a hypothesis, and it is one the codebook cannot settle.

What the corpus does say, in one man's words on 16 August, is that the friction belongs to the light rather than to the dark:

> This is the part of a winter nobody warns you about — the dark was quiet and the coming out is not.

The figure agrees with him: friction runs at two a month through the dark and doubles in September. But that is one observation from one hand, made once, and the codings that support it were made after it was written. It is flagged as an uncoded highlight for that reason.

## The notebook

Mowbray's private book covers the same season and behaves differently under the same codebook, which is the most useful thing in the corpus.

Friction displaced is the code the two documents disagree about. In the journals it is the second most common code in the book, and in the notebook it is nearly absent — not because the leader had no grievances, but because in a book nobody else reads he names people. `callout-5huqdng1` requires the target to be a person the entry reveals rather than states. Mowbray states, so the code does not fire.

That is a definition behaving correctly on prose it was not written for. It also means the code is measuring the presence of an audience as much as the presence of friction, which is worth knowing before anyone reads the September peak as a party falling out.

The saved search **Friction, journal against notebook** puts every friction coding in date order so the two registers can be read against each other.

## What is left open

Eleven passages carry a colour and no code. Three of them are the same shape: a writer noticing a party-wide withholding rather than a particular grievance, in April, May and July. `callout-5huqdng1` covers the particular case and nothing covers the general one. Three instances across three hands is a pattern rather than an accident, and it is the first thing to take to a codebook review.

```json-attributes
{
	"tags": [
		"analysis"
	],
	"date": "1912-11-04",
	"type": "analysis-note",
	"source": "Cape Aldis",
	"subject": "what the codings show across the season",
	"hash": "711ca452860cab28"
}
```
