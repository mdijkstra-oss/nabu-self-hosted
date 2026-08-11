# Codebook — conditions

What the weather, the cold and the light are doing, and what the writer makes of them. These three codes divide a subject the corpus never stops returning to, and the division is the point: the same wind is coded two different ways depending on how the sentence treats it, and the split is only visible to a reader.

```json-callout
{
	"id": "callout-4tfndxpe",
	"type": "codebook-code",
	"title": "Weather as adversary",
	"content": "Weather written as something with intent, capacity or will — an actor the party is up against rather than a condition it is in.\n\nInclusion criteria:\n- Weather given a verb that belongs to an agent: aiming, leaning, refusing, taking, holding out\n- The party positioned against it, as opposing party rather than as observer\n- Weather described by what it does to the hut, the instruments or the men, in terms that would suit an assailant\n\nExclusion criteria:\n- A reading, a figure or a direction recorded without a stance toward it (that is Weather as measurement)\n- The absence of light, its return or its duration (that is Light accounting, even where the sentence resents it)\n- A named man's conduct in bad weather, where the weather is the setting and not the subject\n\nExamples:\n- \"Blowing very hard all day and the hut cracking like a ship.\"\n- \"Two days of it now, the wind off the glacier and into the door as though it had been aimed.\"\n- \"The wind has been from the south for nine days and has stopped being weather.\"\n\nCounter-examples:\n- \"Wind SE force 6 with drift, cleared by evening.\" (a reading, no stance)\n- \"Barometer 29.40 and steady, which is the odd part: the cold has come in quietly and not on a gale.\" (surprise at a figure is not opposition to it)",
	"color": "slate",
	"collapsed": false
}
```

```json-callout
{
	"id": "callout-1zk5b338",
	"type": "codebook-code",
	"title": "Weather as measurement",
	"content": "Weather recorded as quantity: an instrument read, a figure written down, a direction and a force noted.\n\nInclusion criteria:\n- A numeric reading, or a scale value such as a wind force\n- An account of taking, checking or failing to take a reading\n- Commentary on the instrument itself, its error or its limits\n\nExclusion criteria:\n- The same conditions written with a stance toward them (that is Weather as adversary)\n- Hours of sun or minutes of twilight, which are their own code however they are recorded (that is Light accounting)\n\nNote on scope: one entry may carry both this code and Weather as adversary, and several do. That is not an error to be resolved. Where a single sentence does both, code it twice.\n\nExamples:\n- \"Barometer 29.33, wind SE force 3, minimum −38.\"\n- \"Minimum −51, the lowest yet.\"\n- \"Ink useless below about −40 so the figures are all by hand now.\"\n\nCounter-examples:\n- \"Blowing very hard all day and the hut cracking like a ship.\" (conditions, no quantity)\n- \"Sugar fifty-nine.\" (a quantity, but of stores — that is Ration arithmetic)",
	"color": "cyan",
	"collapsed": false
}
```

```json-callout
{
	"id": "callout-6liirel2",
	"type": "codebook-code",
	"title": "Light accounting",
	"content": "Tracking the sun and the twilight: how much there is, how fast it is going or coming, how long it has been absent.\n\nInclusion criteria:\n- Hours of sun, minutes of twilight, days since or until the sun\n- Rates of change in any of those\n- The keeping of the count itself, including keeping it when the value is nil\n\nExclusion criteria:\n- Lamps, blubber and oil, which are stores and not light in this sense\n- Weather readings taken at noon that do not concern the light (that is Weather as measurement)\n\nExamples:\n- \"Twilight at noon thirty-one minutes and lengthening by rather more than three minutes a day, which is faster than my table said and I have checked the table twice.\"\n- \"Because it will not be zero in August, and the day it stops being zero I should like to have the day before it.\"\n- \"Sun above the horizon eleven hours and a few minutes; I shall keep the figure daily, since it is the one quantity here that does what it is told.\"\n\nCounter-examples:\n- \"The lamps are the whole business now.\" (light to work by, not the sun)\n- \"Minimum overnight −47.\" (a reading, but not of light)",
	"color": "amber",
	"collapsed": false
}
```

```json-attributes
{
	"tags": [
		"tag-5wo8ixhn"
	],
	"date": "1912-03-02",
	"type": "codebook",
	"source": "Cape Aldis codebook",
	"subject": "codes for conditions",
	"hash": "66574927b084e7e9"
}
```
