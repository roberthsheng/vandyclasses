# vandyclasses

_I want to learn about **\[X\]**, but YES is slow and unhelpful. What should I take?_

I got annoyed at how slow YES is, how limited their search results were, and the fact that you couldn't search by course description. So I made this. This project was largely inspired by [classes.wtf](https://classes.wtf/), a similar and much better project done for Harvard.

## How does it work?

The main project is written in [Go](https://go.dev/), using data scraped into an in-memory [Redis](https://redis.io/) database. The frontend is mainly written in [HTML](https://developer.mozilla.org/en-US/docs/Web/HTML), [CSS](https://developer.mozilla.org/en-US/docs/Web/CSS), and [JavaScript](https://developer.mozilla.org/en-US/docs/Web/JavaScript), and deployed on [Fly.io](https://fly.io/).

### FAQ

**Why did you make this?** I was bored and had some time on my hands. YES sucks.

**How do you scrape the data?** I used the [Selenium WebDriver](https://www.selenium.dev/documentation/en/webdriver/) to scrape the data from [here](https://www.vanderbilt.edu/catalogs/kuali/undergraduate-23-24.php#/courses). Fun little wormhole of dynamically scraping data.

***Why Go?*** I wanted to learn Go, and it's a really fast systems language while also having low latency.

***Why Redis?*** Per [Eric](https://github.com/ekzhang/classes.wtf/blob/40de058451e33dd60218a3c476cecc69c293fa84/README.md), "it's really fast, it stores data in memory, the API is simple and robust, and it has a best-in-class full-text search module. For this size of dataset, embedding Redis gives you unmatched performance with a fraction of the cost and effort of alternatives".