# Wizarding Exchange API
This is the backend for the wizarding exchange website.

## Design
It's a rather simple API that takes an amount and a currency, and proceeds to convert it into Galleons, Sickles and Knuts.

The conversion base came from this [reddit post](https://www.reddit.com/r/harrypotter/comments/43qv9c/lets_talk_wizard_money_a_look_through_everything/).

The [Fixer API](https://fixer.io/) is used to convert between the base currency(EURO) and the other currencies.

Because the Fixer API provides a limited amount of calls, I decided to also implement a simple "cache" solution. It stores the results from the fixer API into a json file and stores them for up to 1 day. I did look for regular cache options, but they were all a bit too much overkill for what I needed to achieve.

## Purpose
I just wanted to make something in GO and the idea of another TODO list app or blog made my eyes water.


License: MIT