import configparser
import datetime
import json
import random
import requests
import re

config = configparser.ConfigParser()
config.read("config.ini")

username = config["USER"]["Username"]
password = config["USER"]["Password"]

session = requests.Session()

wiki_url = config["WIKI"]["Url"] + "api.php"

def get_token(token_type):
    params = {
        "action": "query",
        "meta": "tokens",
        "type": token_type,
        "format": "json"
    }

    request = session.post(wiki_url, params)

    json = request.json()
    return json["query"]["tokens"][token_type + "token"]

def log_in():
    params = {
        "action": "login",
        "lgname": config["USER"]["Username"],
        "lgpassword": config["USER"]["Password"],
        "lgtoken": get_token("login"),
        "format": "json"
    }

    session.post(wiki_url, params)

def get_article(title):
    params = {
        "action": "query",
        "prop": "revisions",
        "rvprop": "content",
        "rvsection": "0",
        "rvslots": "*",
        "titles": title,
        "format": "json"
    }

    request = session.post(wiki_url, params)

    json = request.json()

    warnings = json.get("warnings", False)
    is_outdated_wiki = True if warnings and warnings["main"]["*"] == "Unrecognized parameter: 'rvslots'" else False

    query_page = json["query"]["pages"]

    if "-1" in query_page:
        # to do: find a better exception class (or make one)
        raise Exception("Page not found")

    first_page = str(min(map(int, query_page.keys())))
    
    if is_outdated_wiki:
        return query_page[first_page]["revisions"][0]["*"]
    
    return query_page[first_page]["revisions"][0]["slots"]["main"]["*"]

def get_facts():
    return [(line, index) for index, line in enumerate(get_article(config["WIKI"]["FactsSourcePage"]).split("\n")) if line.startswith("*") == True]

# these facts have a format like this:
# * text <!--n--> where n is the counter.
# following functions will help parse this as a tuple (text, n)

counter_re = re.compile(r"<!--(?:featured:)?(\d+)(?: time:(\d+))?-->")

def fact_to_tuple(fact, index):
    counter = counter_re.search(fact)

    if counter is None:
        return (fact, 0, index, 0)

    fact_text = fact.replace(counter.group(0), "").rstrip()
    return (fact_text, int(counter.group(1)), index, int(counter.group(2) or 0))

def tuple_to_fact(fact):
    counter = " <!--featured:" + str(fact[1]) + " time:" + str(fact[2]) + "-->" if fact[1] > 0 else ""
    return "%s%s" % (fact[0], counter)

# sample random facts, prioritizing the ones which have been featured the least
def sample_new_facts():
    facts = [fact_to_tuple(f,i) for f,i in get_facts()]

    sample = []

    number_of_facts = int(config["WIKI"]["DidYouKnowFacts"])

    cutoff = 0

    delta = datetime.timedelta(days=int(config["WIKI"]["MinimumDaysBeforeRepeating"]))
    now = datetime.datetime.utcnow()

    while (len(sample) <= number_of_facts):
        minimum_counter_filter = [f for f in facts if f[1] == cutoff and datetime.datetime.utcfromtimestamp(f[3]) < now - delta]

        sample += random.sample(minimum_counter_filter, min(len(minimum_counter_filter), number_of_facts))

        cutoff += 1
        number_of_facts -= len(sample) - 1

    return sample

def edit_page(title, text, summary):
    params = {
        "action": "edit",
        "title": title,
        "text": text,
        "section": 0,
        "summary": summary,
        "bot": "niconiconii",
        "nocreate": "podpá",
        "format": "json",
        "token": get_token("csrf")
    }

    session.post(wiki_url, params)

def update_source_page(facts):
    page = config["WIKI"]["FactsSourcePage"]

    lines = get_article(page).split("\n")

    for fact, counter, index in facts:
        lines[index] = tuple_to_fact((fact, counter, int(datetime.datetime.utcnow().timestamp())))

    edited = "\n".join(lines)

    edit_page(page, edited, "Updating counter for featured trivia facts")

def select_facts():
    facts = sample_new_facts()

    facts_update_counter = [(fact, counter + 1, index) for fact, counter, index, _ in facts]

    update_source_page(facts_update_counter)

    text = "<noinclude><b>This page is edited automatically by the sickest bot you'll ever see, don't mess with it or else you'll die</b></noinclude>\n" + "\n".join([text for text, _, _, _ in facts])

    edit_page(config["WIKI"]["FactsTemplate"], text, "Updating featured facts")

log_in()
select_facts()