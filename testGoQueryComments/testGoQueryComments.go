package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kimxilxyong/intogooglego/post"
	_ "github.com/lib/pq"

	"io"
	"log"
	"os"

	"strconv"
	"strings"
	"time"
	"unicode"
)

func ParseHtmlComments(p *post.Post) (err error) {
	DebugLevel := 3
	// Get comments from hackernews
	//geturl := fmt.Sprintf("http://news.ycombinator.com/item?id=%s", p.WebPostId)
	// DEBUG
	geturl := fmt.Sprintf("https://news.ycombinator.com/item?id=9751858")
	// Get an io.Reader with HTML content
	body := getHtmlInputReader()

	if err != nil {
		return errors.New("GetHtmlBody: " + err.Error())
	}
	// Create a qoquery document to parse from an io.Reader
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return errors.New("Failed to parse HTML: " + err.Error())
	}
	// Find hackernews comments = elements with class "athing"
	thing := doc.Find(".athing")
	for iThing := range thing.Nodes {
		// use `singlecomment` as a selection of one single post
		singlecomment := thing.Eq(iThing)

		comment := post.NewComment()
		//p.Comments = append(p.Comments, &comment)

		comheads := singlecomment.Find(".comhead a")
		for i := range comheads.Nodes {

			comhead := comheads.Eq(i)
			t, _ := comhead.Html()
			s, exists := comhead.Attr("href")
			if exists {
				if strings.HasPrefix(s, "user?id") {
					comment.User = t
					continue
				}
				if strings.HasPrefix(s, "item?id") {
					if strings.Contains(t, "ago") {
						var commentDate time.Time
						commentDate, err = GetDateFromCreatedAgo(t)
						if err != nil {
							comment.Err = errors.New(fmt.Sprintf("Failed to convert to date: %s: %s\n", t, err.Error()))
							err = nil
							continue
						}
						comment.CommentDate = commentDate
						if len(strings.Split(s, "=")) > 1 {
							comment.WebCommentId = strings.Split(s, "=")[1]
						}
						//comment.Err = err
					}
				}
			}

			comments := singlecomment.Find("span.comment").Find("font[color]")
			//comments = singlecomment.Remove() //RemoveClass(".reply")
			var sep string

			fmt.Printf("COMMENT NODES COUNT = %d\n", len(comments.Nodes))
			for iComment, node := range comments.Nodes {
				s := comments.Eq(iComment)
				//comment.Body = comment.Body + sep + stringMinifier(s.Text())
				//h, _ := s.Html()

				n := node.Data

				h := s.Text()

				fmt.Printf("%d - %s: %s\n", iComment, n, h)
				comment.Body = comment.Body + sep + stringMinifier(h)

				sep = "\n"
			}

			fmt.Printf("Body: %s\n", comment.Body)
			//fmt.Printf("COMMENT NODES BODY = %s\n", comment.Body)
			os.Exit(99)
			if comment.Err == nil && len(comment.WebCommentId) > 0 {
				p.Comments = append(p.Comments, &comment)
			}

		}

	}

	if DebugLevel > 2 {
		fmt.Printf("GET COMMENTS FROM %s yielded %d comments\n", geturl, len(p.Comments))
	}

	return err
}

// Removes all unnecessary whitespaces
func stringMinifier(in string) (out string) {

	white := false
	for _, c := range in {
		if unicode.IsSpace(c) {
			if !white {
				out = out + " "
			}
			white = true
		} else {
			out = out + string(c)
			white = false
		}
	}
	return
}

func GetDateFromCreatedAgo(c string) (created time.Time, err error) {

	var amount int64
	var dateunit string
	created = time.Now()

	splitted := strings.Split(c, " ")
	if len(splitted) > 1 {
		amount, err = strconv.ParseInt(splitted[0], 10, 0)
		amount = amount * -1 // Back to the future
		if err != nil {
			err = errors.New(fmt.Sprintf("GetDateFromCreatedAgo: Failed to convert %s: ", c))
			return
		}
		dateunit = splitted[1]
		switch strings.ToLower(dateunit) {
		case "minutes", "minute":
			created = created.Add(time.Duration(amount) * time.Minute)
		case "hours", "hour":
			created = created.Add(time.Duration(amount) * time.Hour)
		case "days", "day":
			created = created.AddDate(0, 0, int(amount))
		case "months", "month":
			created = created.AddDate(0, int(amount), 0)
		case "years", "year":
			created = created.AddDate(int(amount), 0, 0)
		}
	}
	return
}

// returns an io.Reader with dummy test html
func getHtmlInputReader() io.Reader {

	filereader, err := os.Open("testhtmlcomment.html")
	checkErr(err, "Read file failed")
	return filereader

	s := `
<html op="item"><head><meta name="referrer" content="origin"><link rel="stylesheet" type="text/css" href="news.css?s0e3YnyLJyfLknwmR5ca">
        <link rel="shortcut icon" href="favicon.ico">
        <script type="text/javascript">
function hide(id) {
  var el = document.getElementById(id);
  if (el) { el.style.visibility = 'hidden'; }
}
function vote(node) {
  var v = node.id.split(/_/);
  var item = v[1];
  hide('up_'   + item);
  hide('down_' + item);
  var ping = new Image();
  ping.src = node.href;
  return false;
  }
    </script><title>Management things I learned at Imgur | Hacker News</title></head><body><center><table id="hnmain" border="0" cellpadding="0" cellspacing="0" width="85%" bgcolor="#f6f6ef">
        <tr><td bgcolor="#ff6600"><table border="0" cellpadding="0" cellspacing="0" width="100%" style="padding:2px"><tr><td style="width:18px;padding-right:4px"><a href="http://www.ycombinator.com"><img src="y18.gif" width="18" height="18" style="border:1px #ffffff solid;"></a></td>
                  <td style="line-height:12pt; height:10px;"><span class="pagetop">
                              <b><a href="news">Hacker News</a></b><img src="s.gif" height="1" width="10"><a href="newest">new</a> | <a href="newcomments">comments</a> | <a href="show">show</a> | <a href="ask">ask</a> | <a href="jobs">jobs</a> | <a href="submit">submit</a></span></td><td style="text-align:right;padding-right:4px;"><span class="pagetop">
                              <a href="login?goto=item%3Fid%3D9751858">login</a>
                          </span></td>
              </tr></table></td></tr>
<tr style="height:10px"></tr><tr><td>
    <table border="0">
        <tr class='athing'>
      <td align="right" valign="top" class="title"><span class="rank"></span></td>      <td><center><a id="up_9751858" href="vote?for=9751858&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="title"><span class="deadmark"></span><a href="https://medium.com/@gerstenzang/21-management-things-i-learned-at-imgur-7abb72bdf8bf">Management things I learned at Imgur</a><span class="sitebit comhead"> (medium.com)</span></td></tr><tr><td colspan="2"></td><td class="subtext">
        <span class="score" id="score_9751858">202 points</span> by <a href="user?id=jasoncartwright">jasoncartwright</a> <a href="item?id=9751858">8 hours ago</a>                  | <a href="item?id=9751858">133 comments</a>                              </td></tr>
            <tr style="height:10px"></tr><tr><td colspan="2"></td><td><form method="post" action="comment"><input type="hidden" name="parent" value="9751858"><input type="hidden" name="goto" value="item?id=9751858"><input type="hidden" name="hmac" value="041c168187cc7c628cadb2c2139bf8d0d7778259">
    <textarea name="text" rows="6" cols="60"
    style=""
    placeholder=""></textarea><br><br><input type="submit" value="add comment"></form>
</td></tr>
  </table><br><br>
  <table border="0">  <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="0"></td><td valign="top"><center><a id="up_9752282" href="vote?for=9752282&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=mordrax">mordrax</a> <a href="item?id=9752282">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">&quot;It’s terribly difficult to manage unmotivated people. Make your job easier and don’t.&quot;<p>I want to add to this by saying, it&#x27;s terribly difficult to keep yourself motivated in a team with unmotivated people. I left my last job because all the team cared about was money, job security and doing their 9-5.
At my current place, we could fire half the staff and the company wouldn&#x27;t miss a beat. Sadly, there is no management, leadership or clear direction. We are self managed and so it&#x27;s very easy to spot the handful of proactive staff because they usually end up picking up all the tasks which eventually causes burnouts and reduced motivation etc...
I&#x27;ve tried to motivate my team mates but how do you motivate someone who&#x27;s comfortable and secure in their 9-5 and don&#x27;t really care to achieve any more than mediocre? Nobody gets fired and everyone&#x27;s on pretty good pay for the job they do.<p>So I really do agree with this point. Find people who are not satisfied with the status quo. You can&#x27;t change everyone, in fact, you can&#x27;t change anyone. All a manager can do is hope to keep these people accountable if they do not hold themselves responsible.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752282&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752500" href="vote?for=9752500&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=rifung">rifung</a> <a href="item?id=9752500">4 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Is there really anything wrong with the fact that some people don&#x27;t care to achieve anything more than what&#x27;s required of them at their job? I think it&#x27;s unfortunate that you&#x27;re in an environment you don&#x27;t enjoy, and I hope you can find other people you relate with.<p>I hope I don&#x27;t offend you, but it feels a bit selfish to try to change them to be more motivated. For many people, a job is just something they have to get through to survive and there&#x27;s nothing wrong with that. In a way you are lucky because these people will likely never enjoy their jobs as much as you do.<p>Don&#x27;t forget that they may very well have other things besides work that inspire them, but those things might not pay the bills.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752500&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752550" href="vote?for=9752550&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=2chen">2chen</a> <a href="item?id=9752550">3 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">There&#x27;s nothing inherently wrong with that kind of person.  The problem arises when you mix the two types of people in a demanding environment where more than the status quo is expected (e.g., at a startup).  If you want to coast, stick to defense.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752550&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752581" href="vote?for=9752581&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=tensor">tensor</a> <a href="item?id=9752581">3 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">It feels a bit selfish to do the minimum you can at a job because you too afraid to actually try for something you enjoy. Of course, if what you enjoy is doing nothing, then I guess I don&#x27;t have much sympathy.<p>But if you are passionate about making a change in the world, or even just doing an amazing job at one small thing you love, then you are in the right place. Be in the right place. Don&#x27;t treat others and yourself poorly by being in a job you hate.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752581&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752638" href="vote?for=9752638&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=aaronbrethorst">aaronbrethorst</a> <a href="item?id=9752638">2 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000"><p><pre><code>    It feels a bit selfish to do the minimum you
    can at a job because you too afraid to actually
    try for something you enjoy.
</code></pre>
Maybe I have other things outside of work that are far more important to me than trying to ensure that the company that I&#x27;ll have a 0.025% equity stake in <i>in four years</i> becomes the next unicorn. Honestly: why should I care about my founder&#x2F;CEO becoming a multi-millionaire? What&#x27;s in it for me?</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752638&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="160"></td><td valign="top"><center><a id="up_9752813" href="vote?for=9752813&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=mcintyre1994">mcintyre1994</a> <a href="item?id=9752813">1 hour ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I think this is where the mission driven bit comes in. If you don&#x27;t care what the company is doing and their only mission seems to be to make the founders really rich then they&#x27;ve probably screwed up.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752813&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="160"></td><td valign="top"><center><a id="up_9752669" href="vote?for=9752669&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=tdkl">tdkl</a> <a href="item?id=9752669">2 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Job&#x2F;income stability ?</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752669&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="200"></td><td valign="top"><center><a id="up_9752822" href="vote?for=9752822&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=UK-AL">UK-AL</a> <a href="item?id=9752822">1 hour ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">There are a lot better ways of ensuring that that than working yourself silly. Ensuring you have significant savings, skills, connections, being prepared to switch jobs, and being good at negotiations will result in far better results. All of which require time outside of work.<p>Theres nothing wrong fullfilling your agreed work hours, with a reasonable expected work amout and spending time with family.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752822&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="200"></td><td valign="top"><center><a id="up_9752809" href="vote?for=9752809&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=yummybear">yummybear</a> <a href="item?id=9752809">1 hour ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">What makes you think your CEO&#x27;s income is a guarantee for job security?</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752809&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752626" href="vote?for=9752626&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=rifung">rifung</a> <a href="item?id=9752626">3 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Who was asking for your sympathy?</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752626&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752666" href="vote?for=9752666&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=realrocker">realrocker</a> <a href="item?id=9752666">2 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">True there are many who take a job to survive. But I think you are confusing &quot;passion&quot; with motivation. A motivated team member would not sit back while her teammates are passionate about the work.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752666&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752673" href="vote?for=9752673&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=quanticle">quanticle</a> <a href="item?id=9752673">2 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Isn&#x27;t that just a recipe for self-exploitation? I&#x27;m motivated to do my work. But if you discover &quot;passion&quot; and start putting in 80 hours a week, don&#x27;t expect me to run myself into the ground alongside you.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752673&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752530" href="vote?for=9752530&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=nbm">nbm</a> <a href="item?id=9752530">3 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">This may sound weird, but you were probably the one with motivation problems in this scenario, not them.<p>Don&#x27;t believe me?  They&#x27;re comfortable and secure.  They&#x27;re happy with their output.  They&#x27;re apparently paid well.<p>You were unhappy with your situation - you want things to be different than what they are, and it was grating on you (and may or may not have led to burnout from your description), until eventually you left.<p>I&#x27;ve been in a very similar scenario:<p>I joined a &quot;corporate&quot; that had been running an Internet site for a decade.  They had not built a new system in years, and had a massive monolithic ancient app with hundreds of hardcoded rules for individual accounts, where every change required tons of effort and they maintained a QA team at least the size of the dev team.  Very few of the dev team seemed interested in things like learning new technologies, trying out different development methodologies, finding ways to make the system better as opposed to the task list of feature and bug requests.  (I&#x27;m not even sure it had revision control.)<p>Short version - I get annoyed quickly at, for example, rules about arriving at 9:00am (my train schedule meant I tended to arrive at 9:10am, or at 8:30am, and I didn&#x27;t want to hang out at work...).  I&#x27;m told that even though I&#x27;m being productive, I&#x27;m setting a bad example, and they don&#x27;t want others to start doing that.  (I&#x27;m thinking: &quot;Why don&#x27;t you just deal with bad performance, whether the person arrives on time or not?&quot;, as well as &quot;Yeah, and I leave the office at 6pm, so I&#x27;m at the office longer than them!&quot;)<p>Slightly longer version - I work with my manager to get my team (also not happy with the status quo) together in a room somewhat separated from the dev team, and we pump out code for several months a lot happier, until ultimately we all got better offers.<p>There&#x27;s a much longer story, but the point is that _I_ was the one unhappy with the situation, and this affected my motivation to the point of the occasional debilitating day of non-productivity.  The original employees were happy with their situation, and many probably still work there doing the same job they were doing nearly a decade ago now.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752530&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752726" href="vote?for=9752726&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=kabouseng">kabouseng</a> <a href="item?id=9752726">2 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I&#x27;ve been on both sides of the fence, and as you get older you&#x27;ll also probably end up on both sides at times too (No idea what age you are).<p>As everyone mentions sometimes there is other stuff going on in peoples lives, or they are just not that interested in building the same CRUD app for the umpteenth time that has you terribly exited because it is your first go at it and you want to prove yourself.<p>That being said, when on the motivated side of the fence, I want to offer you some advice that was given to me by my manager some years ago: You can&#x27;t keep a good man down. Eventually management will see you put in extra effort, don&#x27;t drop the ball and is constantly pulling the project out of the fire, and that will result in rewards be that promotions or freedoms or whatever. It might not always be as quick as you want, but you can&#x27;t keep a good man down...<p>I hope that helps.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752726&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752536" href="vote?for=9752536&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=adrianhoward">adrianhoward</a> <a href="item?id=9752536">3 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000"><i>&quot;&quot;It’s terribly difficult to manage unmotivated people. Make your job easier and don’t.&quot;</i><p>But it is your job to figure out <i>why</i> you got unmotivated people.<p>* Is your hiring process borked?<p>* Is there something wrong inside your organisation that&#x27;s regularly breaking motivated people?<p>… and so on …</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752536&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752609" href="vote?for=9752609&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=dynamicdispatch">dynamicdispatch</a> <a href="item?id=9752609">3 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Hear, hear.<p>Terrible management and&#x2F;or micromanagement is one of the key reasons a dev can lose motivation. It also doesn&#x27;t do anyone&#x27;s morale&#x2F;motivation any favors to see colleagues get fired. If a manager has had to fire people, and do so repeatedly, then the fault is not so much with those getting fired than the organization&#x2F;management&#x2F;hiring process.<p>One of the key reasons why I&#x27;ve seen dev lose motivation after joining a company is lack of investment on the part of the manager in the employees future goals. I&#x27;ve been in situations during my entry-level days when my manager always made me draw the short straw - did wonders for my motivation.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752609&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752622" href="vote?for=9752622&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=ryandrake">ryandrake</a> <a href="item?id=9752622">3 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I would imagine that people motivated mainly by money are among the easiest to manage. There&#x27;s no guesswork there. Give them clear, measurable performance goals, with dollar signs attached to each one. Done.<p>In my view, money also is a good fallback motivation, if you can&#x27;t figure out exactly what it is that someone wants out of his job, or you can&#x27;t provide it. Truly, I&#x27;ve probably had only one job in my life where I honestly thought &quot;I wouldn&#x27;t stay here NO MATTER WHAT they pay me!&quot;</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752622&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752674" href="vote?for=9752674&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=erikb">erikb</a> <a href="item?id=9752674">2 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I&#x27;d say I&#x27;m not a money oriented person. E.g., looking at the income statistics of my city I think that my chances are &gt;50% if I find any other job.<p>But I think what you say would highly motivate me anyway. People like numbers and making numbers bigger. That&#x27;s why people feel bad if they don&#x27;t get enough Facebook likes. So if you give people clear numbers and a path to increase the numbers they will be motivated to do it, even if they don&#x27;t care what the numbers stand for.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752674&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752783" href="vote?for=9752783&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=cosmolev">cosmolev</a> <a href="item?id=9752783">1 hour ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000"><p><pre><code>  I&#x27;ve probably had only one job in my life where I honestly thought &quot;I wouldn&#x27;t stay here NO MATTER WHAT they pay me!&quot;
</code></pre>
What was this job?</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752783&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752404" href="vote?for=9752404&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=proksoup">proksoup</a> <a href="item?id=9752404">4 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">It&#x27;s shocking to me every time I encounter this. How do those unmotivated people expect to keep their jobs? How do they expect to get the next job, and references for the next job?<p>I&#x27;m less surprised by out of touch management once in a while and more surprised by individuals so confident in their demotivated lifestyle and their ability to maintain it long term with no consequences.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752404&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752842" href="vote?for=9752842&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=UK-AL">UK-AL</a> <a href="item?id=9752842">53 minutes ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Depends what you mean by motivation?<p>I don&#x27;t see anything wrong in working agreed work hours, and completing a reasonable expected work amout and that&#x27;s it. I think that&#x27;s just recognising you have family and other responsibilities.<p>A lot of companies think your not motivated if your not doing 80+ hour weeks, for little added gain. Your not a founder, or have little or no equity and so have a limited upside. I don&#x27;t why companies expect founder level commitment for non founder level upside.<p>In this case your biggest duty is probably to your loved ones and family. And not sacrificing that over some misplaced sense of duty. I&#x27;m sure a lot companies would not do the same if the positions where reversed. They&#x27;d probably fire you pretty fast, if you had issues effecting work. Keep that in mind.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752842&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752437" href="vote?for=9752437&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=DanBC">DanBC</a> <a href="item?id=9752437">4 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">It find it out that the entire blame is put on those people.<p>One or two demotivated people? Well, okay, they&#x27;re probably to blame.<p>But when you have a company full of them it&#x27;s a toxic company and poor management that caused it.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752437&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752476" href="vote?for=9752476&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=adventured">adventured</a> <a href="item?id=9752476">4 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Those people become experts - they get incredibly attuned - at navigating risks when it comes to getting fired. It&#x27;s like people that don&#x27;t work hard consistently, and then cram as much in at the last second right before getting in trouble. They develop an almost sixth-sense like ability to know when they have to do something, and then they do just enough. A lot of times in my observation, persistently unmotivated people don&#x27;t know what they want, either out of work or life in general; so it does them little good to quit and go find a job they want to do, because they have no idea what they want to do.<p>It&#x27;s the classic rhetorical: why not just work consistently, and not have to worry about cramming? Why do people behave that way (in either school or work)? In my personal experience, I&#x27;ve found the only time I cram, is when I&#x27;m stuck doing something I really don&#x27;t care about (happened constantly in school).</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752476&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752743" href="vote?for=9752743&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=odiroot">odiroot</a> <a href="item?id=9752743">1 hour ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Well, good money and job security is also a strong motivator.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752743&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752580" href="vote?for=9752580&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=ljk">ljk</a> <a href="item?id=9752580">3 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">if you don&#x27;t mind me asking, how many hours do you work a week?<p>Entry-level here, and I&#x27;d say i&#x27;m the 9-5 type, but don&#x27;t consider myself &quot;don&#x27;t really care to achieve any more than mediocre&quot; is it possible to be motivated but still work at a reasonable hours? If it&#x27;s not, then I&#x27;m kind of discouraged</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752580&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752584" href="vote?for=9752584&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=tensor">tensor</a> <a href="item?id=9752584">3 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Yes, absolutely. Everyone should be paid for the time they put in. Being motivated means being great at what you do, not working for free.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752584&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="0"></td><td valign="top"><center><a id="up_9752076" href="vote?for=9752076&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=kzhahou">kzhahou</a> <a href="item?id=9752076">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">When did it become a Universal Truth that good management requires regular scheduled 1-1s (weekly or biweekly at most)?  Clearly managers should meet frequently and in depth with people on the team, but the &quot;regular 1-on-1&quot; thing is mantra, and I&#x27;m not sure there&#x27;s research&#x2F;evidence to support its efficacy.<p>People never look forward to them.  You gotta remember everything you did that week so you can report on it (even when told the 1-1 isn&#x27;t meant for status reports... it kinda always is).  You gotta think of some issue to bring up to manager&#x27;s attention.  It interrupts your whole day.  Your manager didn&#x27;t yet help you on your problem from last week -- now they want to listen thoughtfully to problems this week?<p>From the manager&#x27;s POV: You&#x27;ve been keeping up with the team members all week, helping out and checking in every day... now you&#x27;re gonna lose one or two days with back-to-back 1-1s which leave your voice hoarse.  They need a few more days to finish out whatever you discussed last week... should you cancel the 1-1 and sync up on the next week (and lose the 1-1 tempo), or have a &quot;no-diff&quot; meeting that you both know you coulda skipped?</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752076&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752089" href="vote?for=9752089&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=Blackthorn">Blackthorn</a> <a href="item?id=9752089">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">&gt; People never look forward to them.<p>I&#x27;m sorry, but...speak for yourself. I have nothing but respect for my manager and he respects me as well. It&#x27;s there so if I have some persistent problem or concern with our project&#x27;s direction, they can help. And, yes, it does provide the opportunity for you to market yourself and what you&#x27;ve done. Marketing is important.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752089&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752288" href="vote?for=9752288&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=dmak">dmak</a> <a href="item?id=9752288">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I also look forward to 1 on 1s. In fact, I have to be the one to schedule and ask for them. It&#x27;s a good way to build a communication channel that allows constructive feedback. Communication is always good.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752288&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752096" href="vote?for=9752096&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=PurplePanda">PurplePanda</a> <a href="item?id=9752096">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">If I had some persistent problem or concern I would just bring it up, I wouldn&#x27;t wait for some scheduled meeting. I guess some managers might not be so easily available.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752096&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752305" href="vote?for=9752305&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=bikesandcode">bikesandcode</a> <a href="item?id=9752305">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">It&#x27;s not just the managers. I&#x27;ve worked with great managers who were truly available, but due how busy I perceived everyone to be, I never felt comfortable interrupting things for what I deemed weren&#x27;t critical issues.<p>Since we did have regular 1 on 1 meetings, I knew I had a dedicated block of time coming up to hash things out.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752305&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752273" href="vote?for=9752273&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=balls187">balls187</a> <a href="item?id=9752273">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Or some employees don&#x27;t feel comfortable bringing it up.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752273&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752201" href="vote?for=9752201&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=hueving">hueving</a> <a href="item?id=9752201">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">If your manager has no idea what your value is without your marketing, you work in a shitty environment.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752201&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752573" href="vote?for=9752573&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=RyanZAG">RyanZAG</a> <a href="item?id=9752573">3 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Some people like to feel appreciated, so they&#x27;d talk up how well they&#x27;re doing during the 1:1 each week and glow happily as their manager praises them for their good job - while yawning himself to sleep inside. Then the employee goes off feeling &#x27;wanted&#x27; and motivated and everyone is happy.<p>Obviously this wouldn&#x27;t be useful for everyone, but it&#x27;s useful for some people. They&#x27;re usually pretty easy to spot, too.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752573&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752172" href="vote?for=9752172&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=vidarh">vidarh</a> <a href="item?id=9752172">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">The regular 1-on-1 thing is mantra because people are notoriously bad at determining when they are actually <i>needed</i>, and because so many managers spend far less time actually listening to their reports than they think. They often spend lots of time <i>with</i> their reports, but usually talking to them or over them, or focusing entirely on tasks rather than giving attention to the person.<p>If 1-1s take one or two days of back to back meetings, to me that means the team is either too large for you to effectively manage without putting putting in place a team lead or more to delegate to. Or alternatively those meetings are <i>really</i> necessary, or you wouldn&#x27;t have that much to talk about. If it leaves someones voice hoarse as a manager, it makes me think they&#x27;re not listening enough.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752172&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752243" href="vote?for=9752243&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=kzhahou">kzhahou</a> <a href="item?id=9752243">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">&gt; the team is either too large<p>Not really.  It can be tough even at 8-10 reports, 30 minutes each (some may go over): 4-5 hours of meetings.  It&#x27;s hard to schedule them literally back-to-back because of everyone else&#x27;s schedules, so it winds up taking up two afternoons or whatever (it won&#x27;t <i>literally</i> take up two full days from morning to evening).<p>&gt; If it leaves someones voice hoarse as a manager, it makes me think they&#x27;re not listening enough.<p>Very cute.  Maybe I don&#x27;t have the stamina for non-stop talking as others do :-P</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752243&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752890" href="vote?for=9752890&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=vidarh">vidarh</a> <a href="item?id=9752890">26 minutes ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">8-10 directs reports is already too large in my opinion, at least to me, exactly because of the amount of effort it takes to follow all of them up sufficiently.  I&#x27;ve had teams with that many direct reports, and it <i>sucks</i> for me and it sucks for the people reporting to me if there&#x27;s not then a team lead or similar exactly because it drains time. I don&#x27;t see cutting the time I spend on people as viable at all.<p>The threshold might very well differ from person to person, but in my opinion, if you can&#x27;t afford to schedule 30 minutes per person on a weekly basis (whether or not you actually end up using the full amount of time) without feeling it is too much, then you can&#x27;t afford to manage that many people directly, no matter how many people we&#x27;re talking. For some that might mean 10 or even more works, but in my experience as the team size grows the amount of time you need to invest per person grows as well as interpersonal issues and communications gets more complicated.<p>&gt; Very cute.<p>Maybe, but it&#x27;s not meant to be - I&#x27;m very serious on that point. When I do 1-on-1&#x27;s with reports, if I&#x27;m the one doing the talking it&#x27;s a sign we&#x27;ve spent more time than necessary and the meeting is at an end.<p>That might also be part of the reason why I have a different attitude to them: If a report doesn&#x27;t have anything to bring up, the meeting is over in 5 minutes. But in my experience at least, there&#x27;s a big difference in what comes out when you bring someone into a meeting room for 5 minutes (or on a private call), and specifically <i>ask them</i> in a one on one setting if there&#x27;s anything they&#x27;d like to bring up, anything I could help with, what personal development they&#x27;d like doing etc..<p>I&#x27;m sure that for some people that&#x27;s not necessary and they get people to open up without creating that setting all the time, but my experience with my own managers too is that managers in IT (probably applies elsewhere too, but I&#x27;ve only ever worked in tech companies) are notoriously bad at creating good environments for this. And I need to keep a very close eye on myself too, as it&#x27;s not a part of the job that comes naturally to me.<p><i>That</i> is where the ceremony comes it. It&#x27;s not for the people who are great at getting everyone to talk. It&#x27;s for everyone else - including a lot of people who <i>think</i> they&#x27;re great at getting people to talk.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752890&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752474" href="vote?for=9752474&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=foz">foz</a> <a href="item?id=9752474">4 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">As a manager, 1-1s should not just be considered status reports. They are a way for you to build trust and communication with your staff, and to give them a chance to talk about anything: problems, work on personal development, or feedback about you as their manager.<p>1-1s are probably your most valuable tool. Often, you will discover problems and info you would have otherwise not known - &quot;our lead developer seems depressed and was talking about quitting&quot;, &quot;did you hear about that other project that started? It&#x27;s in direct conflict with our plans...&quot;, &quot;there&#x27;s a conference coming up, we should present at it&quot;, and so forth.<p>An effective manager should, in my opinion, spend 50% or more of his time with his team. Working on the same topics, talking to them, helping to plan, fixing problems, finding resources, and doing 1-1s. It&#x27;s no surprise that teams with the most problems often have a manager who is just not around enough.<p>For new managers, I always suggest the following resources as a great starting point:<p>- &quot;Team Geek&quot;: <a href="http:&#x2F;&#x2F;shop.oreilly.com&#x2F;product&#x2F;0636920018025.do" rel="nofollow">http:&#x2F;&#x2F;shop.oreilly.com&#x2F;product&#x2F;0636920018025.do</a>
- &quot;Managing Humans&quot;: <a href="http:&#x2F;&#x2F;www.amazon.com&#x2F;Managing-Humans-Humorous-Software-Engineering&#x2F;dp&#x2F;1430243147" rel="nofollow">http:&#x2F;&#x2F;www.amazon.com&#x2F;Managing-Humans-Humorous-Software-Engi...</a>
- Manager Tools podcast: <a href="https:&#x2F;&#x2F;www.manager-tools.com" rel="nofollow">https:&#x2F;&#x2F;www.manager-tools.com</a></font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752474&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752690" href="vote?for=9752690&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=quanticle">quanticle</a> <a href="item?id=9752690">2 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">&gt;1-1s should not just be considered status reports.<p>And scrum standups should only take 5 minutes. Theory is one thing. Reality is another. It&#x27;s even worse when you have a team that does scrum, and also schedules regular 1:1s with the boss. I&#x27;m expected to give my status every day, and then summarize a week&#x27;s worth of status at the 1:1 meeting. It&#x27;s a colossal waste of time, and I don&#x27;t know how to &quot;take charge&quot; of my 1:1 (or even if I&#x27;m supposed to take charge) in order to refocus it onto what I want it to be about.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752690&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752787" href="vote?for=9752787&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=foz">foz</a> <a href="item?id=9752787">1 hour ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">When 1-1s become the weekly status report, your manager has failed you. Your manager should know what&#x27;s up because he is around, uses the team&#x27;s tools, and asks questions.<p>You should definitely take charge of 1-1s. Maybe try bringing a list of non-status related issues and hand them over. The problems you see, ideas you have, off-topic things. Make it  clear to your boss that things aren&#x27;t working.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752787&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752484" href="vote?for=9752484&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=rifung">rifung</a> <a href="item?id=9752484">4 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Clearly it&#x27;s not a universal truth because at least from what you say, it sounds like you don&#x27;t like them.<p>On the other hand, I&#x27;ll have to say you&#x27;re definitely wrong to say people don&#x27;t look forward to them. I have yet to find a company where we have had too many 1:1s, and I usually have them weekly.<p>It&#x27;s true that sometimes I don&#x27;t have anything I really need to communicate, and what ends up happening is that the meeting takes 5 minutes and that&#x27;s it. That might seem like a waste, but I like it because it makes me feel like my manager actually cares enough to hear the feedback from us underlings. If we didn&#x27;t have scheduled 1:1s, I would have a really hard time bringing up anything because my manager is very often elsewhere.<p>I also feel like it should be beneficial from the manager&#x27;s perspective because assuming people actually are comfortable talking about issues, there are usually problems which arise that the manager doesn&#x27;t have to deal with and thus know about. Then, we can make sure these things get addressed.<p>Of course, all this is moot if you don&#x27;t feel like you can trust your manager, which is unfortunately sometimes the case.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752484&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752268" href="vote?for=9752268&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=balls187">balls187</a> <a href="item?id=9752268">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">FWIW, my career took a turn for much better when I started working at a place that had regular 1 on 1&#x27;s.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752268&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752193" href="vote?for=9752193&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=baddox">baddox</a> <a href="item?id=9752193">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I think of regularly scheduled 1-on-1s as a shortcut to make employees feel like management is open to serious communication. It&#x27;s a regularly-scheduled explicit opportunity to give feedback, so even if no specific feedback is usually necessary, when it <i>is</i> necessary the time slot is already booked.<p>Obviously, being open to communication is a good thing, so I&#x27;m not really critical of 1-on-1s. The fact that there isn&#x27;t usually anything substance to talk about is a feature, not a bug. Sure, ideally the company culture and the individual relationships with managers would be such that everyone knows that serious discussions can be initiated any time they&#x27;re necessary, but failing that, regularly-scheduled 1-on-1s are probably the next best thing.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752193&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752211" href="vote?for=9752211&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=prostoalex">prostoalex</a> <a href="item?id=9752211">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">It&#x27;s a venue to bring up problems that could be bottled up. Properly conducted 1-on-1&#x27;s don&#x27;t have to be dreadful and can end quickly if so desired.<p>Manager can probe with questions such as &quot;What&#x27;s the biggest obstacle you&#x27;re facing right now?&quot; and &quot;Anything you need from me to help you out?&quot;, but there are some weeks where nothing will come up, which would end the session.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752211&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752369" href="vote?for=9752369&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=ryandrake">ryandrake</a> <a href="item?id=9752369">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">The help&#x2F;harm of 1-on-1s depends on the manager&#x27;s style. I&#x27;ve had managers whose 1:1s were helpful and motivating, and others where it felt like I was on trial each time, and I dreaded seeing that calendar notification pop up.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752369&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752462" href="vote?for=9752462&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=DanBC">DanBC</a> <a href="item?id=9752462">4 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Some people like &quot;reflective practice&quot; and &quot;supervision&quot;. (Here supervision is a jargon word that comes from healthcare or protective social services. A person has regular meetings with a more experienced team member to talk about difficult things they&#x27;ve had that week &#x2F; month.)<p>The challenges that programmers face are very different to the challenges a mental health nurse faces, but they&#x27;re still useful concepts. Weekly might be a bit much for programmers.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752462&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752094" href="vote?for=9752094&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=thawkins">thawkins</a> <a href="item?id=9752094">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Agreed, i dont do 1:1&#x27;s with my team a) i have about 90 folks to deal with, so it would be near to impossible. b) all my folks know (and do) that they can grab 30 mins at any time with me if they have something on thier mind. I kinda keep track of who is not talking to me, and hit them up with a chat now and then. 
I also do something I call &quot;walking the floor&quot;, where every few hours i get around everybody and just say hi and ask how everybody is doing, and what they are up to, ask if they have any problems. Show interest in what they are doing. I usualy time it when i know folks are takimg natural breaks etc so i know im not distracting them.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752094&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752258" href="vote?for=9752258&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=codeonfire">codeonfire</a> <a href="item?id=9752258">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">&gt;I also do something I call &quot;walking the floor&quot;<p>Also known as drive-by management.  Why not schedule some time instead?</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752258&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752308" href="vote?for=9752308&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=jsmthrowaway">jsmthrowaway</a> <a href="item?id=9752308">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Among &quot;worst qualities of management I&#x27;ve ever experienced,&quot; drive-by is in the top 5. Even satirized in <i>Office Space</i>.<p>Come to think of it, &quot;far too many direct reports&quot; is on the same list...</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752308&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752192" href="vote?for=9752192&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=serge2k">serge2k</a> <a href="item?id=9752192">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">90 direct reports?</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752192&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752106" href="vote?for=9752106&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=yeukhon">yeukhon</a> <a href="item?id=9752106">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">It seems like too much to have one person managing all 90ds. I guess this is where people either agree or disagree about the layers of middle management...for example have local leads reports to local manager, and local manager reports back to director&#x2F;VP&#x2F;SVP or whatever if you have people concentrated in various location.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752106&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752643" href="vote?for=9752643&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=DannyBee">DannyBee</a> <a href="item?id=9752643">2 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I try to keep my org as flat as possible, but 90 people would be completely nutso.<p>To me it&#x27;s a sign of an org that can&#x27;t grow managers properly.<p>To put this in perspective: if he gave each of his folks just 26 minutes a week, that would be every minute of working time in a 40 hour work week :)</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752643&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752642" href="vote?for=9752642&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=aaronbrethorst">aaronbrethorst</a> <a href="item?id=9752642">2 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Where do you work?</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752642&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752160" href="vote?for=9752160&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=majormajor">majormajor</a> <a href="item?id=9752160">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I look forward to 1-on-1s with my manager(s) and I&#x27;ve never used them for status reports.<p>Perfect time to ask questions about things not directly in my day to day domain.<p>Weekly seems a bit much for anyone but super-new-hires, though.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752160&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752695" href="vote?for=9752695&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=Kiro">Kiro</a> <a href="item?id=9752695">2 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#888888">&gt; People never look forward to them<p>100% not true. Explain and apologize.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752695&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="0"></td><td valign="top"><center><a id="up_9752153" href="vote?for=9752153&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=nbm">nbm</a> <a href="item?id=9752153">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">If one considers the primary purpose of managers to be to improve the performance of those that report to them, then this list starts poorly.<p>It _is_ terribly difficult to manage unmotivated people, but most people only end up unmotivated because of poor management.  And it&#x27;s usually cheaper and ultimately better to motivate a proved achiever who has lower motivation than to find a replacement.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752153&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752232" href="vote?for=9752232&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=nostrademons">nostrademons</a> <a href="item?id=9752232">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">That&#x27;s not always true.  Sometimes people end up unmotivated because the organization shifts its mission in one direction while they shift their values in another, and so what starts out as great mission-alignment falls out of alignment.  Sometimes people grow out of a role when the role doesn&#x27;t grow with them.  Sometimes peoples&#x27; life circumstances change and they don&#x27;t have as much attention to devote to work.  Sometimes people are just looking to collect a paycheck and don&#x27;t care about the work at all.<p>A manager should always <i>start</i> by assuming that a motivation problem is something he can address, and work with the employee to figure out what the real reason is and see if they can make the necessary changes.  But sometimes, the employee and organization just aren&#x27;t a good fit anymore, and it&#x27;s better for both of them if they part ways and find new situations that <i>are</i> a good fit.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752232&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752285" href="vote?for=9752285&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=nbm">nbm</a> <a href="item?id=9752285">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Absolutely.  There are many other ways people end up unmotivated, but the most common I&#x27;ve seen has to do with things that are within the realm of their manager - possibly with escalation to someone higher up.<p>I&#x27;ve encountered only five managers who fully exemplify this, but they had an outsize role in keeping good people who others might not have been able to keep, keeping them productive (at least enough of the time to satisfy their work commitments while they dealt with other issues), and attracting more good people (because of word of mouth from the existing good people).  It seems the key here was to have the conversations when they were needed.  The advice here &quot;Make your job easier and don’t.&quot; doesn&#x27;t correlate with these great managers at all.  At least two of them to my knowledge have given reports the advice (and assistance) to find opportunities elsewhere.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752285&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752279" href="vote?for=9752279&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=seanmcdirmid">seanmcdirmid</a> <a href="item?id=9752279">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Isn&#x27;t that just management at a higher level? You can have a great direct manager, but you are still at the mercy of his&#x2F;her managers, and the re-orgs, mission shifts, and everything else. I guess this is why big companies often prefer employees who are motivated by things other than the work they do (like by the paycheck, as you mention). Problem is, you can&#x27;t really be very innovative with that.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752279&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752198" href="vote?for=9752198&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=danieltillett">danieltillett</a> <a href="item?id=9752198">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">The most unmotivated people I have ever had the displeasure of having to manage were people I had no ability to fire or promote. There is basically nothing you can do yet higher management still holds you responsible for the lack of results.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752198&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752239" href="vote?for=9752239&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=nbm">nbm</a> <a href="item?id=9752239">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I&#x27;ve been on the receiving side of &quot;no ability to promote&quot;, and ultimately the best thing a manager can do in that situation is to have an honest conversation about what is and isn&#x27;t possible.<p>In addition, don&#x27;t just give up.  Keep talking and helping out.  It could be advice on how to stay motivated or at least not demotivated, carving out space for interesting projects or developing new skills, or even helping that person find a new opportunity outside the company.<p>Dealing with unmotivated unproductive people you have limited ability to actually manage is beyond my experience.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752239&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752798" href="vote?for=9752798&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=danieltillett">danieltillett</a> <a href="item?id=9752798">1 hour ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">It is most definitely not an experience you want. The best you can do is flow with it. The thing that gets tiring very quickly is having to do their work as well as your own because otherwise innocent people get harmed.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752798&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752260" href="vote?for=9752260&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=balls187">balls187</a> <a href="item?id=9752260">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#9c9c9c">&gt; In addition, don&#x27;t just give up. Keep talking and helping out.<p>Motivation is the responsibility of the employee. As a manager, if your performance is faltering and it&#x27;s because you&#x27;re not motivated, I will tell you clearly my perspective, and ask if that is the case. However it&#x27;s on you take corrective action. I will do everything I can to help you, but it&#x27;s your responsibility, and not fair to the other team members who are motivated if<p>1. I am focusing my energy dealing with issues caused by your lack of motivation.<p>2. They&#x27;re working extra to ensure success when a team member is unmotivated.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752260&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752219" href="vote?for=9752219&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=skeeterbug">skeeterbug</a> <a href="item?id=9752219">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">If you can&#x27;t fire or promote them are you really their manager?</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752219&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752794" href="vote?for=9752794&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=danieltillett">danieltillett</a> <a href="item?id=9752794">1 hour ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Not really, but when you are held responsible for their &quot;productivity&quot; (or more accurately lack of productivity) you have all the responsibility and none of the resources to manage.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752794&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752264" href="vote?for=9752264&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=balls187">balls187</a> <a href="item?id=9752264">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Even if you are in an &quot;At-Will&quot; employment state, it&#x27;s not as simple to fire an employee. You have to have HR involved, and prepare enough documentation to satisfy the need to terminate someone.<p>In smaller orgs with no HR, it&#x27;s probably easier, but larger companies have a specific process.<p>Though to your point, I would not consider myself someones manager unless I had the option to hire, fire, or promote them.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752264&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752271" href="vote?for=9752271&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=yeukhon">yeukhon</a> <a href="item?id=9752271">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Union, tenure, or simply no from upper management. Sometimes people deliver the result but they may behave in a way that shows they are not motivated. That is a hard thing to use to fire someone.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752271&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="160"></td><td valign="top"><center><a id="up_9752801" href="vote?for=9752801&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=danieltillett">danieltillett</a> <a href="item?id=9752801">1 hour ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Exactly. My experience is from within the university system where nobody can be fired for non-performance. You get people that are just waiting to retire and don&#x27;t care about anyone other than themselves.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752801&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
          <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752216" href="vote?for=9752216&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=mentat">mentat</a> <a href="item?id=9752216">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">&gt; most people only end up unmotivated because of poor management
I don&#x27;t find this to be a self evident truth. Can you provide some argumentation around this?</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752216&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752253" href="vote?for=9752253&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=balls187">balls187</a> <a href="item?id=9752253">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Most people who <i>leave</i> their jobs do so and cite management.<p>Whether or not that has to do with their motivation is unclear.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752253&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="0"></td><td valign="top"><center><a id="up_9752116" href="vote?for=9752116&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=lohengramm">lohengramm</a> <a href="item?id=9752116">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">&quot;It’s terribly difficult to manage unmotivated people. Make your job easier and don’t.&quot;<p>And where is the challenge if everyone is highly motivated and easy to manage? Then you can just leave the programmers alone. They don&#x27;t need management in this case.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752116&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752182" href="vote?for=9752182&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=pan69">pan69</a> <a href="item?id=9752182">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Most programmers I know (myself included) don&#x27;t need managers but facilitators, people who get shit out of the way so I can do my job.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752182&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752132" href="vote?for=9752132&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=wcrichton">wcrichton</a> <a href="item?id=9752132">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Just because people are highly motivated doesn&#x27;t mean they&#x27;re intrinsically organized. You can leave the programmers alone, but you still need someone to decide what they&#x27;re working on and how to best allocate resources. A manager ensures a team stays coordinated.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752132&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752173" href="vote?for=9752173&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=nbm">nbm</a> <a href="item?id=9752173">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I&#x27;ve found that groups of individual contributors are actually very good on deciding what is important, what they should work on and when, and how to work together to get things done.<p>Managers can provide a useful role in setting the scene, reminding what the goals of the organisation, the department, and so forth are, and how this connects with what the team is doing.  They are perhaps most useful in working with individuals who aren&#x27;t working well as part of the team at the moment - giving them feedback, mentorship, building bridges, connecting them with people, training, and so forth.<p>In my opinion, managers hurt more than not if they override the self-organisation of teams, and this most hurts when you need two teams to work together - and one of the managers has (explicitly, perhaps) made it clear that doing things &quot;off the plan&quot; is not appreciated.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752173&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752294" href="vote?for=9752294&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=jsmthrowaway">jsmthrowaway</a> <a href="item?id=9752294">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">That <i>strongly strongly strongly</i> depends on the team and contributors.<p>I&#x27;d agree at my startup employers. I would not agree about several teams I worked with at Apple, to pick on a corporate example. Teams built from the ashes of an acquired startup at Apple, again, I&#x27;d be more inclined to agree. You can&#x27;t support a broad conclusion like that anecdotally, because I can counterexample it anecdotally, implying there&#x27;s more to it.<p>It really depends on the ICs in question. Startups are far more selective about their ICs because one person has a very big impact. With the exception of Google and a couple others, large-cap corporate throws IC quantity at problems and distinguished, autonomous, &quot;rockstar&quot; (sigh) ICs are far more rare. You need the cat herders there.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752294&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="160"></td><td valign="top"><center><a id="up_9752333" href="vote?for=9752333&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=nbm">nbm</a> <a href="item?id=9752333">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">It does depend on the contributors, and obviously depends on the culture of the company (ie, if everyone assumes managers are supposed to do it, nobody will do it).<p>You can&#x27;t rely on any random grouping of people to decide well on what&#x27;s important to do and how to effectively break up the work.  But adding a random manager to that group doesn&#x27;t help specifically.  Adding a more experienced IC will generally help more than a less experienced manager.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752333&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="160"></td><td valign="top"><center><img src="s.gif" height="1" width="14"></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  [deleted]              </span><div class='reply'></div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="200"></td><td valign="top"><center><a id="up_9752411" href="vote?for=9752411&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=1123581321">1123581321</a> <a href="item?id=9752411">4 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Individual Contributor.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752411&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
          <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752204" href="vote?for=9752204&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=lohengramm">lohengramm</a> <a href="item?id=9752204">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I find it hard to believe that a complete highly motivated team would need such a guidance. Unless we are not thinking about the same thing when we use the word &quot;motivated&quot;.<p>To me, a complete highly motivated team (a team without any unmotivated person) would certainly feel like they should bring up some kind of organization. They are motivated, after all, and thus they want to get their work done.<p>A not motivated person, however, needs to be pushed, not in a bad way, but in some way, because he&#x2F;she is unmotivated (don&#x27;t feel like wanting to get any job done), and thus the hierarchy comes into play.<p>Most people fall in this &quot;unmotivated&quot; category. And that is (I think) the actual true reason why management exists in the real world: to push (unmotivated) people, so they get their work done.<p>Note: by &quot;push&quot; I don&#x27;t mean &quot;being an idiot&quot;. I believe that the presence of the manager is already sufficient for most people to feel like they should work, even though they don&#x27;t want to.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752204&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752632" href="vote?for=9752632&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=bitmage">bitmage</a> <a href="item?id=9752632">2 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">There&#x27;s another critical role in many organizations for your manager; defending your team against other managers.
Otherwise your highly-motivated group will become less so as they are blamed for failures beyond their control, have their schedules randomly changed for &#x27;urgent&#x27; work, etc.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752632&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752318" href="vote?for=9752318&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=balls187">balls187</a> <a href="item?id=9752318">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I look at managing difficult situations as a challenge, but I have to be certain that I believe this is the right thing for the person, team, and ultimately company--not just a way to prove how awesome of a leader I am.<p>I didn&#x27;t get the sense the author was advocating just dumping someone on the grounds they have a lack of motivation, but rather understand that we have the ability to terminate someone and that is sometimes the best course for everyone involved--including the person who is struggling with performance.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752318&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752129" href="vote?for=9752129&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=mosquito242">mosquito242</a> <a href="item?id=9752129">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">management at that point becomes providing direction and helping focus priorities and helping your programmers be their best.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752129&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="0"></td><td valign="top"><center><a id="up_9752825" href="vote?for=9752825&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=datashovel">datashovel</a> <a href="item?id=9752825">1 hour ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">The tone of the list seemed to draw a line between the &quot;manager&quot; and the &quot;managed&quot;.<p>Personally I think the best way to manage a team is to draw a line between the &quot;team&quot; (including management) and the &quot;objectives&quot;.<p>The seemingly tough part with doing things this way is to be able to draw a line between individuals and the rest of the team if someone isn&#x27;t holding up their weight.<p>It sounds tough, but I think it doesn&#x27;t have to be as tough as it sounds.  If people are given enough independence and individual responsibilities, but in a collaborative setting, it will become obvious to the entire team when individuals are not carrying their weight.<p>Keep in mind that once a person becomes a member of your team you&#x27;re in it for the long haul.  Every once in a while a team needs to rally to achieve objectives, since it&#x27;s almost impossible to give perfectly equal distribution of work in a project.  Every once in a while individuals need to be propped up by the rest of the team if they&#x27;re having difficulties with their individual responsibilities.<p>It should only become a problem when the same individuals, over the course of an extended period of time, fail to meet their objectives.  And even then if the company is large enough they should be given an opportunity to move laterally within the company to move to another team.  If individuals fail regularly within multiple teams at that point it becomes obvious they may need to be let go.<p>With all that said, different projects &#x2F; companies have different budgets and so they may need to cut corners.  That&#x27;s why it&#x27;s important to be careful to select people you are in it for the long haul with, and are willing to put the time and energy into that person as an investment for the betterment of the company.<p>Sudden turnover is one of the worst things for morale in any company.  Avoid it at just about any cost, except in extraordinary circumstances.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752825&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="0"></td><td valign="top"><center><a id="up_9752278" href="vote?for=9752278&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=balls187">balls187</a> <a href="item?id=9752278">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Good list, but what this hits on is managing down, and not much on how to manage up, or manage across.<p>In my experience, I&#x27;ve found managing a team fairly easy--I just think about all the things my great managers did, and try to emulate them.<p>Managing up, and managing other teams however, has been a real learning experience.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752278&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="0"></td><td valign="top"><center><a id="up_9752473" href="vote?for=9752473&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=zongitsrinzler">zongitsrinzler</a> <a href="item?id=9752473">4 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">These 21 points make Imgur seem like an over-managed hell pit to work at.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752473&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752497" href="vote?for=9752497&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=prawn">prawn</a> <a href="item?id=9752497">4 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I think that if <i>good</i> management were following these rules, you wouldn&#x27;t even know it was happening. It&#x27;d be more like guiding and less like controlling.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752497&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="0"></td><td valign="top"><center><a id="up_9752596" href="vote?for=9752596&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=ryandrake">ryandrake</a> <a href="item?id=9752596">3 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">&gt; Define clear merit-based systems, which reduces confusion about what your team members need to do be recognized.<p>This is a great one, I wish more companies&#x2F;managers did this. It&#x27;s terrible when your company&#x27;s advancement policy is &quot;if your manager like you, you get promoted&quot;. It&#x27;s demotivating when your bonus is based on someone&#x27;s subjective feeling about how good a job you&#x27;re doing, or some ridiculous self-assessment essay.<p>Give me clear, measurable goals, and a clear, scheduled (on the calendar) performance review.<p>Ship Product A or complete features B and C on time and your next raise will be X.
Ship it on time and under budget and your bonus will be Y.
Otherwise, Z% cost of living increase only.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752596&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752795" href="vote?for=9752795&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=cactusface">cactusface</a> <a href="item?id=9752795">1 hour ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">What if all your time is spent cleaning up after the cowboys?  Or talking to the other team members about the problems they are stuck working on?  A mission critical 2-line bugfix can take months.  There aren&#x27;t objective measures of performance.  Any metric you come up with, employees will find a way to thwart it.  Shipping a product on time?  Easy.  Just don&#x27;t mention the bugs.  Surely you don&#x27;t believe it&#x27;s possible to create bug-free software, do you?  When all the team members have their performance tied to them, they&#x27;ll all collude to game metrics.  Don&#x27;t even get me started on SLOC.  Speaking from experience here, sorry if it&#x27;s too cynical &#x2F; bitter.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752795&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752792" href="vote?for=9752792&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=skrebbel">skrebbel</a> <a href="item?id=9752792">1 hour ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Nearly everything i care about in a great programmer is way less well measurable than &quot;ship feature X on time&quot;. I like your sentiment but I think real good work (for any office job, really) is very hard to measure.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752792&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="0"></td><td valign="top"><center><a id="up_9752821" href="vote?for=9752821&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=AriaMinaei">AriaMinaei</a> <a href="item?id=9752821">1 hour ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">&gt;&gt; <i>Define success clearly and don’t flip-flop on the definition without new information.</i><p>Does anyone have an example for this?</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752821&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="0"></td><td valign="top"><center><a id="up_9752343" href="vote?for=9752343&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=caseyf7">caseyf7</a> <a href="item?id=9752343">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Didn&#x27;t Imgur have a total of 13 people last year?</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752343&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752452" href="vote?for=9752452&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=mburst">mburst</a> <a href="item?id=9752452">4 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Yea we&#x27;ve grown like crazy over the past year. We&#x27;re now up to around 55 people.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752452&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752576" href="vote?for=9752576&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=nbm">nbm</a> <a href="item?id=9752576">3 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">My experience is that interesting things happen to organisations at size 20-30 (basically when the CEO and&#x2F;or COO can no longer keep tabs on nearly everything themselves, no longer can maintain some relationship with all staff), and also when there are more than 33% of people with &lt;6 months at the company.  Hope this wasn&#x27;t as painful as some of my experiences with just single dosages of these at a time.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752576&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="0"></td><td valign="top"><center><a id="up_9752485" href="vote?for=9752485&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=mandeepj">mandeepj</a> <a href="item?id=9752485">4 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">&gt; Give feedback frequently and directly. As a manager, it’s easier to wait and then hedge critical feedback in soft wrappers, but that’s selfish. I’d try to give feedback as soon as I could grab a conference room with the person, and not wait until the formal 1:1 days later.<p>I really wish if everyone could think like this. What is the point in giving feedback after 6 or 12 months during appraisal. That is also done as part of formality.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752485&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="0"></td><td valign="top"><center><a id="up_9752901" href="vote?for=9752901&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=lifeisstillgood">lifeisstillgood</a> <a href="item?id=9752901">19 minutes ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">&gt; People need to feel like they’ve been listened to, not to make the final call. Take the time to listen (you might be wrong), make a decision and then explain the decision. Don’t offer commentary on others’ decisions until you understand why the decisions were made.<p>very true</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752901&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="0"></td><td valign="top"><center><a id="up_9752196" href="vote?for=9752196&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=codinghorror">codinghorror</a> <a href="item?id=9752196">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Hmm. Sam was only at Imgur for less than a year. That&#x27;s a lot of management lessons for a limited period of time...</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752196&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752559" href="vote?for=9752559&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=BinaryIdiot">BinaryIdiot</a> <a href="item?id=9752559">3 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">You&#x27;d be surprised with how many lessons you can take away from even a short period of taking a new position.<p>For instance I had done very limited management in my past. When a new project came up the company I worked at they had made me lead the project which essentially meant developer + manager. This employer also had the habit of under staffing teams while pushing very aggressive timelines. This lead me to try to manage a team of 6 while also coding around the clock for several months. When it came to review time I was a mess; I hadn&#x27;t paid enough attention to my team, I was constantly scrambling for tasks, trying to find ways to help where necessary, etc. It was an incredible trial-by-fire.<p>When it came to my second stint at the same position for a new project I was ready. I gave constant feedback in real-time, I was able to keep on top of tasks, my meetings were more streamlined and I was able to estimate things far better; I did a 180 and became much more effective.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752559&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752370" href="vote?for=9752370&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=taw-snark0">taw-snark0</a> <a href="item?id=9752370">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">He writes well, but is only 2 years out of college (Stanford &#x27;13). Learning that (and noticing how he obfuscated it on his LinkedIn) made me take this piece with an even larger grain of salt than I would normally have.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752370&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752608" href="vote?for=9752608&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=nathankleyn">nathankleyn</a> <a href="item?id=9752608">3 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">He probably obfuscated that fact so people wouldn&#x27;t judge him just by his age; it would seem based on some of the comments in this thread that he was right to do so.<p>Some of the best lessons in life can come from those who haven&#x27;t been tainted by long stints doing the very thing they&#x27;re commenting on.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752608&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752587" href="vote?for=9752587&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=dynamicdispatch">dynamicdispatch</a> <a href="item?id=9752587">3 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">It&#x27;s also interesting to note that the company itself has grown from 13-55 people within that year. You&#x27;d imagine they&#x27;d have someone with a bit more experience managing the team (unless it&#x27;s a team where the median age is, say, 23). Inexperienced managers are a bigger risk than inexperienced devs, and if I were at Imagur, I&#x27;d be wary of (and maybe also unmotivated by) a young manager trying to prove a point (or feature on the front page of HN).</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752587&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752302" href="vote?for=9752302&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=balls187">balls187</a> <a href="item?id=9752302">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Maybe he was a quick study?</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752302&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="0"></td><td valign="top"><center><a id="up_9752146" href="vote?for=9752146&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=danieltillett">danieltillett</a> <a href="item?id=9752146">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Number 6 was the hardest for me to learn - I always wanted to give people another chance to change, but they never did. It is always hard to fire someone, but once you have reached the decision to fire do it as soon as possible.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752146&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="0"></td><td valign="top"><center><a id="up_9752760" href="vote?for=9752760&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=known">known</a> <a href="item?id=9752760">1 hour ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Management inversely proportional to adversity &amp; diversity</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752760&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="0"></td><td valign="top"><center><a id="up_9752468" href="vote?for=9752468&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=spotman">spotman</a> <a href="item?id=9752468">4 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I like a lot of these very much and agree. Thanks for posting.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752468&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="0"></td><td valign="top"><center><a id="up_9752055" href="vote?for=9752055&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=jqm">jqm</a> <a href="item?id=9752055">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I thought this was a good article. I believe most people who are worth having want to be part of something bigger themselves, want to contribute to the world and are willing to sacrifice a bit to do so. These are the people to identify and reward and it is always frustrating when this doesn&#x27;t happen.<p>Per brobdingnagian&#x27;s comment... everyone has bad days on occasion. A good manager can work around this. But a workplace is about work. Not therapy.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752055&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752517" href="vote?for=9752517&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=rifung">rifung</a> <a href="item?id=9752517">3 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I guess I&#x27;m just still too young, naive, and perhaps stupid, but I disagree with this idea strongly. It certainly is true that most places (in the US at least) treat work like work and not like therapy.<p>Maybe you&#x27;re right and work is about work and not therapy, but why not? If anything it seems like it&#x27;s in the company&#x27;s best interest to provide therapy if it&#x27;s needed. They already provide health insurance, and it&#x27;s well known that people who are happier perform better on the job than people who aren&#x27;t.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752517&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752680" href="vote?for=9752680&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=jqm">jqm</a> <a href="item?id=9752680">2 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Remember work for you is product for someone else.<p>All the things we take for granted... the lights being on, the roads being paved, the bus coming on time... all this happens because people organize, take care of their personal affairs, show up and perform their duties in a proscribed manner, even when they would rather be doing something else.<p>My belief is to play when it&#x27;s time to play and work when it&#x27;s time to work and to keep my personal life separate from work and not to turn personal problems into workplace problems. Managers aren&#x27;t (generally) personal counselors nor should they be. Their function in that role is to insure something happens in accordance with organizational goals. Not to be someone&#x27;s mom or life coach. And again... the function of the workplace is to provide a good or a service to someone else. Not a place for working out personal issues. Take that to the appropriate venue lest the buses stop coming on time.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752680&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752067" href="vote?for=9752067&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=brobdingnagian">brobdingnagian</a> <a href="item?id=9752067">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">That&#x27;s really a false dichotomy. Ideally you aren&#x27;t a nice person outside of work and a jerk at work. You are just a nice person. And if you see someone struggling, you don&#x27;t fire them: you help them.<p>Just like most of the best practices here haven&#x27;t been tied to success, it&#x27;s also true that helping people going through a hard time hasn&#x27;t been tied to failure.<p>The right thing to do is really obvious here. Help them.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752067&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752117" href="vote?for=9752117&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=technofiend">technofiend</a> <a href="item?id=9752117">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Well you say that but I work at a very high pressure job that requires me to be far harsher towards employees than friends.<p>I test as ENTP at work and INTJ at home just because success at work requires a different mindset.  I don&#x27;t like it, but that&#x27;s how it is.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752117&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752154" href="vote?for=9752154&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=wtvanhest">wtvanhest</a> <a href="item?id=9752154">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">The last company I worked for was extremely high pressure, but I never encountered anyone who was harsh.  They would be fired.  It just isn&#x27;t necessary.<p>Also, ENTP &amp; INTJ all that stuff is proven to be bunk with no statistical evidence that it predicts anything.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752154&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="160"></td><td valign="top"><center><a id="up_9752431" href="vote?for=9752431&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=technofiend">technofiend</a> <a href="item?id=9752431">4 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Well at work I&#x27;m forced to give much more honest and direct feedback than I may do outside of work.<p>I&#x27;m also required to represent my company&#x27;s interest instead of lending a sympathetic ear to a friend. Don&#x27;t get me wrong: I&#x27;m honest and fair, but some avenues are simply closed in an employee &#x2F; employer relationship.<p>I just told a guy today that he flunked an interview because he failed to convince the ED he interviewed with that he (the interviewee) would adequately represent his area on a P1 production support bridge.<p>I could have sugar-coated it and said &quot;Ah, man, yeah that was unfair because I&#x27;ve seen you succeed in exactly that situation&quot; but instead I coached him on how to present himself better and (hopefully) pass the interview next time.<p>It wasn&#x27;t pleasant telling him what I had to tell him, but it was necessary.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752431&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752141" href="vote?for=9752141&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=brobdingnagian">brobdingnagian</a> <a href="item?id=9752141">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">&quot;I don&#x27;t like it, but that&#x27;s how it is.&quot;<p>No arguing with that.<p>In the long run, we can hopefully select against business practices that have these kinds of human costs.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752141&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
          <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="0"></td><td valign="top"><center><a id="up_9752008" href="vote?for=9752008&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=brobdingnagian">brobdingnagian</a> <a href="item?id=9752008">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#bebebe">Right off the bat you tip your hand: you are selecting against people who are depressed. Rather than try to help them, you&#x27;d rather let them wallow in misery, fail, become unemployed, schizophrenic and eventually die homeless.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752008&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752047" href="vote?for=9752047&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=pen2l">pen2l</a> <a href="item?id=9752047">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">That&#x27;s horribly off-topic, but I&#x27;ll continue being horribly off-topic, only because I had a very strong realization about this today.<p>My boss has been bugging me to get stuff done since there are some crazy deadlines coming up for us. I realized a few days ago that were it not for the craziness at work, I would be very lonely... I have nothing to do. And it&#x27;s precisely when I have nothing to do that I start having negative or suicidal thoughts. So having a job really is a blessing for me, it gives me something to keep my mind preoccupied, and gives me an easy opportunity to socialize (walking up to someone and starting a conversation with someone is a lot easier.. when... well, you have to, because work requires that you have to :)).</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752047&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752050" href="vote?for=9752050&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=brobdingnagian">brobdingnagian</a> <a href="item?id=9752050">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Thanks for sharing. I don&#x27;t think it&#x27;s off topic.<p>&quot;It’s terribly difficult to manage unmotivated people. Make your job easier and don’t.&quot;<p>It&#x27;s just a really really mean thing to say, especially coming from someone who is a manager &#x2F; mentor.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752050&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752073" href="vote?for=9752073&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=jordigh">jordigh</a> <a href="item?id=9752073">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I understand trying to be humane to everyone, but you cannot just decide to not fire anyone, ever. If someone does not want to be at a job, it is not an act of charity to keep them at that job. If they really do not want to be there, they can cause a lot more grief to themselves or others. It&#x27;s cruel to be kind.<p>You&#x27;re not Schindler. Firing people is not a death sentence. If you fear for their mental health, fire them and direct them to services that can help. Hell, companies even do this routinely in a fashion: it&#x27;s called outplacement.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752073&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="160"></td><td valign="top"><center><a id="up_9752156" href="vote?for=9752156&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=smacktoward">smacktoward</a> <a href="item?id=9752156">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000"><i>&gt;  If you fear for their mental health, fire them and direct them to services that can help.</i><p>If we&#x27;re talking about Americans, their health insurance is almost certainly tied to their job. So firing them will force them to choose to either pay out the nose for COBRA or individual health insurance -- which means more stress, on top of the stress of losing their job -- or to give up access to the vast majority of those &quot;services that can help&quot; altogether.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752156&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="200"></td><td valign="top"><center><a id="up_9752263" href="vote?for=9752263&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=jordigh">jordigh</a> <a href="item?id=9752263">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#737373">&gt; If we&#x27;re talking about Americans,<p>Oops, sorry, I forgot. I&#x27;m in Canada.<p>I guess it is inhumane to fire people in the US, then.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752263&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="160"></td><td valign="top"><center><a id="up_9752247" href="vote?for=9752247&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=nether">nether</a> <a href="item?id=9752247">6 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I wonder if it&#x27;s illegal to fire someone for being depressed per the American Disabilities Act, since it protects those with &quot;mental impairments&quot;: <a href="http:&#x2F;&#x2F;www2.nami.org&#x2F;Template.cfm?Section=Helpline1&amp;template=&#x2F;ContentManagement&#x2F;ContentDisplay.cfm&amp;ContentID=4862" rel="nofollow">http:&#x2F;&#x2F;www2.nami.org&#x2F;Template.cfm?Section=Helpline1&amp;template...</a></font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752247&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="200"></td><td valign="top"><center><a id="up_9752654" href="vote?for=9752654&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=DannyBee">DannyBee</a> <a href="item?id=9752654">2 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">It&#x27;s a complex topic, but the way you phrase it, the answer is &quot;yes&quot;.
You cannot fire someone <i>because they have depression</i> (assuming a real diagnosis, etc)
The real question is can you fire someone <i>who performs poorly because of their depression</i></font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752654&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752066" href="vote?for=9752066&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=pen2l">pen2l</a> <a href="item?id=9752066">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">I agree that&#x27;s a shitty thing. Also, sadly, it is a norm to think in this bleakly capitalistic way. By and large, you will not see normal people viewing the providing job&#x2F;no job thing as something moral.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752066&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752084" href="vote?for=9752084&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=oh_sigh">oh_sigh</a> <a href="item?id=9752084">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Not to be heartless, but why would it be a businesses concern if someone was depressed or not?</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752084&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752862" href="vote?for=9752862&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=Sideloader">Sideloader</a> <a href="item?id=9752862">41 minutes ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">When a person says &quot;Not to be x, but&quot;, being x is exactly what they are doing. It&#x27;s wilful self-delusion. Either fully own and accept your decision or, if that&#x27;s problematic, consider that maybe it&#x27;s not a good decision and consider an alternative course of action.<p>If a CEO&#x2F;hiring manager decides employees that become depressed, get sick, are involved in an accident, whatever are to be shown the door rather than offered help or assistance, fine, that is their prerogative. But don&#x27;t pretend it&#x27;s not a shitty way to treat a person.<p>And employers wonder why so many people absolutely despise their jobs. Treating a person like a piece of equipment that can and should be replaced immediately if a part malfunctions or it isn&#x27;t performing optimally is an excellent way to ensure an oppressive and hostile work environment. But then you get to crack the whip and scare some &quot;motivation&quot; into the workforce, and maybe that&#x27;s exactly what you wanted to do. Just don&#x27;t expect a grateful and productive staff.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752862&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752312" href="vote?for=9752312&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=Sgt_Apone">Sgt_Apone</a> <a href="item?id=9752312">5 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Depression in the workforce causes reduced productivity and substantial amounts of money on a national level. The CDC estimations that it causes 200 million lost workdays each year which costs employers between $17 to $44 billion alone. It&#x27;s in an organization&#x27;s interest to have healthy employees; both mentally and physically.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752312&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752922" href="vote?for=9752922&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=Sideloader">Sideloader</a> <a href="item?id=9752922">8 minutes ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Pfffft...come on, what&#x27;s this socialist obsession with facts and logic? Just fire some people already. And do something about that bleeding heart!</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752922&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752093" href="vote?for=9752093&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=brobdingnagian">brobdingnagian</a> <a href="item?id=9752093">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Let&#x27;s do a reductio-style thought experiment: what if everyone was depressed? That&#x27;s obviously a business concern.<p>The standard argument for socialism follows from the same reductio: what if everyone was happy? That would be GREAT for business.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752093&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752916" href="vote?for=9752916&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=Sideloader">Sideloader</a> <a href="item?id=9752916">11 minutes ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#5a5a5a">You are an idiot and I am so fucking glad I never have to work under assholes like you ever again.<p>Oh, nice straw man you&#x27;ve built. Because treating employees like human beings = slippery slope to socialism.<p>The level of fail in American management culture boggles the mind. These sad fools don&#x27;t get that employee productivity and, thus, the company&#x27;s earnings, would improve if they treated people with basic respect. An employee who actually <i>wants</i> to be at work is much more productive than the person who dreads coming to work every day. A lot of managers are assholes just because they can be and they should be fired on sight because their incompetence drags the entire team down and saddles the company with a shitty reputation. A miserable demotivated workforce plus a bad reputation in the industry (all of which could easily have been avoided) is like regularly shovelling bundles of $100 bills into the basement incinerator. Sadly, many companies shovel many bundles into many fires all over the nation and then promote the lead shoveller so he can start burning $1000 bundles.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752916&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
    <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752498" href="vote?for=9752498&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=adventured">adventured</a> <a href="item?id=9752498">4 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">The second one is not true. Utopia of happiness would result in vast stagnation.<p>Agitation, annoyance, something needing fixed, imperfection, desire for more &#x2F; self-improvement, etc. is a critical root of invention and creativity.<p>Ideally not everyone is perfectly happy all at the same time. And it should be noted that is not the same as saying that everyone should never be happy. Rather, that it&#x27;s incredibly valuable to have dissatisfied people in society - they&#x27;re often the ones that break with the status quo and push humanity forward.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752498&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="40"></td><td valign="top"><center><a id="up_9752024" href="vote?for=9752024&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=cbd1984">cbd1984</a> <a href="item?id=9752024">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Depression causes schizophrenia?</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752024&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="80"></td><td valign="top"><center><a id="up_9752029" href="vote?for=9752029&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=brobdingnagian">brobdingnagian</a> <a href="item?id=9752029">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#aeaeae">I wouldn&#x27;t say causes - but the statistics speak volumes. A quarter of homeless people are schizophrenic.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752029&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752048" href="vote?for=9752048&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=jordigh">jordigh</a> <a href="item?id=9752048">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">That&#x27;s because the schizophrenics were basically evicted from the psychiatric hospitals. With nowhere to go and no one to care for them, they ended up homeless.<p><a href="https:&#x2F;&#x2F;en.wikipedia.org&#x2F;wiki&#x2F;Homelessness_and_mental_health#Deinstitutionalization" rel="nofollow">https:&#x2F;&#x2F;en.wikipedia.org&#x2F;wiki&#x2F;Homelessness_and_mental_health...</a><p>Also, how did you manage to relate this at all to TFA?</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752048&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="160"></td><td valign="top"><center><a id="up_9752058" href="vote?for=9752058&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=brobdingnagian">brobdingnagian</a> <a href="item?id=9752058">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#888888">That accounts for some of the homeless schizophrenics.<p>&quot;Also, how did you manage to relate this at all to TFA?&quot;<p>Just what exactly do you think the real world implications of this attitude &#x2F; policy are more generally?<p>&quot;It’s terribly difficult to manage unmotivated people. Make your job easier and don’t.&quot;<p>He is saying: don&#x27;t try to manage them. Fire them.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752058&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
      <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="120"></td><td valign="top"><center><a id="up_9752039" href="vote?for=9752039&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=minimaxir">minimaxir</a> <a href="item?id=9752039">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#000000">Correlation does not imply causation.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752039&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
        <tr class='athing'><td><table border="0">  <tr><td class='ind'><img src="s.gif" height="1" width="160"></td><td valign="top"><center><a id="up_9752044" href="vote?for=9752044&amp;dir=up&amp;goto=item%3Fid%3D9751858"><div class="votearrow" title="upvote"></div></a></center></td><td class="default"><div style="margin-top:2px; margin-bottom:-10px;"><span class="comhead">
          <a href="user?id=brobdingnagian">brobdingnagian</a> <a href="item?id=9752044">7 hours ago</a> <span class="par"></span><span class="deadmark"></span>          <span class='storyon'></span>
                  </span></div><br><span class="comment">
                  <font color="#9c9c9c">Have enough heart to acknowledge what&#x27;s going on in front of you.</font>
              </span><div class='reply'>        <p><font size="1">
                      <u><a href="reply?id=9752044&amp;goto=item%3Fid%3D9751858">reply</a></u>
                  </font>
      </div></td></tr>
      </table></td></tr>
          </table><br><br>
  </td></tr>
<tr><td><img src="s.gif" height="10" width="0"><table width="100%" cellspacing="0" cellpadding="1"><tr><td bgcolor="#ff6600"></td></tr></table><br><center><span class="yclinks">
                <a href="newsguidelines.html">Guidelines</a>
        | <a href="newsfaq.html">FAQ</a>
        | <a href="mailto:hn@ycombinator.com">Support</a>
        | <a href="https://github.com/HackerNews/API">API</a>
        | <a href="security.html">Security</a>
        | <a href="lists">Lists</a>
        | <a href="bookmarklet.html">Bookmarklet</a>
        | <a href="dmca.html">DMCA</a>
        | <a href="http://www.ycombinator.com/apply/">Apply to YC</a>
        | <a href="mailto:hn@ycombinator.com">Contact</a></span><br><br>
                <form method="get" action="//hn.algolia.com/">Search:
          <input type="text" name="q" value="" size="17"></form>
                    </center></td></tr>      </table></center></body></html>
`
	return strings.NewReader(s)
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

func main() {
	post := post.Post{}
	ParseHtmlComments(&post)
}
