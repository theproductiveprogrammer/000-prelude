/**
*** Hey there! Are you a programmer
*** who is interested in
*** becoming more productive? Welcome!
*/
/**
    welcome.png
*/
/**
*** This post is just the "prelude",
*** where I create the blogging engine
*** to publish my blogs. So this post
*** won't interest you (unless you
*** want to create a similar blog)
*** Instead, you should head out and
*** check [href=/](my later posts).
*/
/**
    bye.png [href=/]
*/
/**
*** Still here? Ok then, let's dive
*** in.
*/
/** A fair question to ask is
*** just _why_ (as a so-called
*** "productive programmer"),
*** would I write a blogging
*** engine rather than just
*** use the dozens of options
*** available -
*** from the amazingly powerful
*** [href=https://wordpress.org/](WordPress) to the light and
*** speedy [href=https://jekyllrb.com/](Jekyll).
*/
/**
*** It's just silly...isn't it?
*/
/**
*** Well I agree with you - it
*** _is_ pretty dumb. Except for
*** two things:
***  1. I want each blog post
***     to not just be me rambling
***     about theoretical ideas,
***     but to produce something
***     of actual value that works.
***  2. If I _am_ a productive
***     programmer then - heck -
***     it shouldn't take me more
***     than a few hours to create
***     a simple blogging engine.
***     Right?
*/
/**
*** And because it shouldn't take
*** more than a few fun hours to
*** create *and* it lets me create
*** a blog which delivers actual
*** working project with every
*** post (this first blog will be
*** a downloadable and usable
*** engine by itself), I've decided
*** to just go ahead and build it.
*/

/**
    done.png [href=https://gitlab.com/productiveprogrammer/000-prelude/blob/master/prelude.go]
*/

/** [...]
I'm done with the engine! It's now
usable and (as you can see) is
generating the blog you are
reading.
It took a bit longer than expected
but not too much and I'm quite
happy with the way it turned
out. What follows is the code
itself that is part of this
blog post and can be found
in [href=https://gitlab.com/productiveprogrammer/000-prelude/blob/master/prelude.go](this file).
*/

/**
[=] So what do we have to do?
We'll take an input file with
paths to the blog posts and use
it to:
(a) Generate a first/index page
with links to
(b) Generated blog posts
Simple enough? Let's begin...
*/
package main

import (
    "os"
    "log"
    "io/ioutil"
    "strings"
    "regexp"
    "time"
    "errors"
    "os/exec"
    "path/filepath"
    "text/template"
    "html"
    "strconv"
    "fmt"
)

func main() {

    postinfo,err := load_config()
    if err != nil {
        log.Fatal(err);
    }

    postinfo,err = set_post_info(postinfo)
    if err != nil {
        log.Fatal(err);
    }

    generate_blog_index(postinfo)
    generate_blog_posts(postinfo)
}

var LINE_MARKER string = "[\n\r]+"
var WHITESPACE string = "[ \t]"

/*
[:typ:]
*/
type post_comment_marker struct {
    start       string
    decorate    rune
    end         string
}
type postcontent_type int
const (
    EMPTY postcontent_type = iota
    POSTCOMMENT
    CODE
)
type PostContent struct {
    Typ     postcontent_type
    HTMLVal string
}
type PostInfo struct {
    InPath  string
    On      time.Time
    Content []PostContent

    pcm     post_comment_marker

    OutPath string
    HTMLTitle   string
}

/**
* [...]
* The config file simply
* contains / the list of paths
* to each blog post
*     posts/timemgmt/timemgmt.c
*     posts/learn-angular/angularstart.htm
*     ...
*/
func load_config() ([]PostInfo,error) {
    cfg, err := get_config_file()
    if err != nil {
        return nil,err
    }
    var data []byte
    data, err = ioutil.ReadFile(cfg)

    lines := regexp.MustCompile(LINE_MARKER).Split(string(data), -1)

    var r []PostInfo
    for _,line := range lines {
        line = strings.TrimSpace(line)
        if len(line) > 0 {
            r = append(r, PostInfo{ InPath: filepath.Clean(line) })
        }
    }
    return r,nil
}

/*[=] Get the config file from
* the user */
func get_config_file() (string,
error) {
    if len(os.Args) == 1 {
        return "", errors.New("No config file provided!")
    }
    return os.Args[1], nil
}

/**
[=] Set the information we need
for our posts from the source
file. This means:
[ ] Set the post date
[ ] Set the post content
[ ] Set the post out path
[ ] Set the post out title
*/
func set_post_info(pi []PostInfo) ([]PostInfo,error) {
    var err error
    for ndx := range pi {

        pi[ndx].On,err = get_post_date(pi[ndx])
        if err != nil {
            return nil,err
        }

        pi[ndx].Content,err = get_post_content(pi[ndx])
        if err != nil {
            return nil,err
        }

        pi[ndx].OutPath,err = get_outpath(pi[ndx])
        if err != nil {
            return nil,err
        }

        pi[ndx].HTMLTitle,pi[ndx].Content,err = get_post_title(pi[ndx])
        if err != nil {
            return nil,err
        }

    }
    return pi,nil
}

/*
[=] Return the output path
for the post.
[ ] The is the filename + ".php"
*/
func get_outpath(postinfo PostInfo) (string,error) {
    return filepath.Base(postinfo.InPath) + ".php" ,nil
}

/*
[:cond:]
*/
func cond_is_title(c PostContent) bool {
    v := strings.TrimSpace(c.HTMLVal)
    return c.Typ == POSTCOMMENT && len(v) > 0 && !strings.Contains(v,"\n")
}

/*
[=] Return the title of the post
[ ] If the first content is a
POSTCOMMENT with only one line
we use that as the title.
[ ] Otherwise we use the file
name as the title (replacing
underscores with spaces)
*/
func get_post_title(postinfo PostInfo) (string,[]PostContent,error) {
    if len(postinfo.Content) > 0 && cond_is_title(postinfo.Content[0]) {
        return postinfo.Content[0].HTMLVal,postinfo.Content[1:],nil
    }

    return fname_to_title(filepath.Base(postinfo.InPath)),postinfo.Content,nil
}


/*
  [=] Convert file name to a
  title-like string
*/
func fname_to_title(fname string) string {
    return template.HTMLEscapeString(strings.Replace(fname, "_", " ", -1))
}


/**
[!] The post date is not
explicitly set. And because the
post repositories are replicated
across dev and production, they
do not share a date. Therefore
setting a post date can be a
little tricky.
[+] We first try to get a date
from git. This is not perfect as
git doesn't track file date so
we use the latest commit
information as a proxy.
[+ -] When starting a new post,
the file is not in git and does
not contain commit information.
So we default to file
modification time as a fallback.
*/
func get_post_date(postinfo PostInfo) (time.Time,error) {
    var t time.Time

    out,err := exec.Command("git", "log", "-n1", "--format=%ad", "--date=short", "--", postinfo.InPath).Output()
    date := strings.TrimSpace(string(out))
    if err == nil && len(date) > 0 {
        t,err = time.Parse("2006-01-02", date)
        if err != nil {
            return t, errors.New("Failed to parse git date: " + date)
        }
        return t,nil
    }

    var fi os.FileInfo
    fi,err = os.Stat(postinfo.InPath)
    if err != nil {
        return t,err
    }
    return fi.ModTime(),nil
}

/*
[=] Use the file data to create
post content of different types.
The steps we follow are:
[ ] Find "post block comment
marker" for this type of file.
    - For example:
        .js files   : /** * /
        .htm files  : <!---- -->
        .nim files  : ## ##
        ...
[ ] Read the file data and convert
into post blocks and code blocks.
[ ] The blocks are processed and
returned.
*/
func get_post_content(postinfo PostInfo) ([]PostContent,error) {
    var err error

    postinfo.pcm, err = get_comment_marker(postinfo)
    if err != nil {
        return nil, err
    }

    var data []byte
    data, err = ioutil.ReadFile(postinfo.InPath)
    if err != nil {
        return nil, err
    }

    return process_post_content(split_post_content(data, postinfo), postinfo), nil
}

/**
[=] Split the file data into post
content
[=] The kind of splitting we have
to do differs if we have a line
type commment:
        ## This starting marker
        ## matches the
        ## ending marker so the
        ## block ends when the
        ## marker is missing
Or a block type comment:
        /** This starting marker
        ** does not match the
        ** ending marker so
        ** the block ends when
        ** the ending marker is
        ** found * /
[ ] Check what type of block we
have and split appropriatly.
*/
func split_post_content(data []byte, postinfo PostInfo) []PostContent {

    cond_is_line_type_block := func(postinfo PostInfo) bool {
        return postinfo.pcm.start == postinfo.pcm.end
    }

    if cond_is_line_type_block(postinfo) {
        return split_post_content_linecomments(data, postinfo)
    } else {
        return split_post_content_blockcomments(data, postinfo)
    }

}
/**
[=] Split file based on block-type
post comments
[ ] Convert the data to a string and
add guards on both ends so that
we can match regular expressions
that start with newline without
worrying about edge cases.
[ ] Loop finding post block
comment marker start
[ ] All content till the start
marker is put into a CODE block
[ ] Close the block by finding
a line that matches the ending
marker.
[ ] The content of this block is
put as a POSTCOMMENT block and the
loop is continued.
*/
func split_post_content_blockcomments(data []byte, postinfo PostInfo) []PostContent {
    var r []PostContent

    rx_start := regexp.MustCompile(LINE_MARKER + regexp.QuoteMeta(postinfo.pcm.start))
    rx_end := regexp.MustCompile(regexp.QuoteMeta(postinfo.pcm.end))

    content := "\n" + string(data)

    for {
        m_start := rx_start.FindStringIndex(content)
        if m_start == nil {
            r = append(r, PostContent{ Typ: CODE, HTMLVal: content })
            return r;
        }
        if m_start[0] > 0 {
            r = append(r, PostContent{ Typ: CODE, HTMLVal: content[:m_start[0]] })
        }
        content = content[m_start[1]:]

        m_end := rx_end.FindStringIndex(content)
        if m_end == nil {
            r = append(r, PostContent{ Typ: POSTCOMMENT, HTMLVal: content })
            return r;
        }
        r = append(r, PostContent{ Typ: POSTCOMMENT, HTMLVal: content[:m_end[0]] })
        content = content[m_end[1]:]
    }
}

/**
[=] Split file based on line-type
post comments
[ ] Split the content into lines
[ ] Start with an accumulator
of "empty line" type
[ ] While the current line is
of the same type, continue to
accumulate it.
[ ] If the current line is of
a different type, add a new
record of the existing accumulator
and start a new accumulator
of the new type
[ ] When all lines are over,
create a record of the remaining
accumulator
*/
func split_post_content_linecomments(data []byte, postinfo PostInfo) []PostContent {
    var r []PostContent

    rx := regexp.MustCompile(regexp.QuoteMeta(postinfo.pcm.start))
    rx_line_ending := regexp.MustCompile("\n|\r|\n\r|\r\n")

    lines := rx_line_ending.Split(string(data), -1)

    content_type := func(line string) postcontent_type {
        line = strings.TrimSpace(line)
        if len(line) == 0 {
            return EMPTY
        }
        if rx.FindStringIndex(line) != nil {
            return POSTCOMMENT
        }
        return CODE
    }

    type accum_ struct {
        typ postcontent_type
        cnt []string
    }

    accum := accum_{}

    accum_lines := func(line string) {
        accum.cnt = append(accum.cnt, line)
    }
    empty_accum := func(typ postcontent_type) {
        if accum.typ != EMPTY {
            r = append(r, PostContent{ Typ: accum.typ, HTMLVal: "\n" + strings.Join(accum.cnt,"\n") })
        }
        accum.typ = typ
        accum.cnt = []string{}
    }

    for _,line := range lines {
        typ := content_type(line)
        if typ != accum.typ {
            empty_accum(typ)
        }
        accum_lines(line)
    }
    empty_accum(EMPTY)

    return r
}

/**
[!] The post content contains
markup-like text I would like to use:
    [href=.](link text)
    https://www. youtube.com/watch?v=XXXXXX
    some_pic .png
    *bold*
    _italic_
    *_class1_*
    *__class2__*
    *___class3___*
    ...
[!] The content also contains
text that needs to be escaped
in order to form valid HTML
(like <, >, &, etc...)
[+] Escape the content of all
text, look for remaining patterns
and replace with the appropriate
HTML. ie:
[ ] First we clean the post content
of any decorators.
[ ] Escape HTML for all blocks
[ ] If the block is not POSTCOMMENT
    type, we're done
[ ] Otherwise, find the relevant
markup and replace it.
*/
func process_post_content(pcs []PostContent, postinfo PostInfo) []PostContent {
    var r []PostContent

    for _,pc := range pcs {
        pc.HTMLVal = clean_post_content(pc, postinfo.pcm.decorate)
        pc.HTMLVal = template.HTMLEscapeString(pc.HTMLVal)
        if pc.Typ == POSTCOMMENT {
            pc.HTMLVal = replace_markup(pc.HTMLVal, postinfo)
        }
        r = append(r, pc)
    }

    return r
}

/**
[=] Post content sometimes
contain decorators:
    /** Some
    *** Text with
    *** Deocorators * /
which we need to clean up
*/
func clean_post_content(pc PostContent, decorater rune) string {

    rx := regexp.MustCompile(LINE_MARKER + regexp.QuoteMeta(string(decorater)) + "+" + WHITESPACE + "?")

    if pc.Typ == POSTCOMMENT {
        return rx.ReplaceAllString(pc.HTMLVal, "\n")
    } else {
        return pc.HTMLVal
    }
}

/**
[=] Replace all markup within
the content.
    [href=.](link text)
    https://www. youtube.com/watch?v=XXXXXX
    some_pic .png
    some_pic .png [href=.]
    *bold*
    _italic_
    *_class1_*
    *__class2__*
    *___class3___*
    ...
[+] Find the appropriate regular
expressions, and replace them.
[+ -] The tricky bit is to not replace
expressions that contain URL's. For
example:
    href=/the/_best_/part
    should *NOT* become
    href=/the/<i>best</i>/part
[+] So what we'll do is save
the url's in an array and temporarily
index them by using $$$$<num>$$$$, which
(hopefully) should never be found in
our text.
[ ] Find all matches starting with URL
matches (so we can safetly save them away).
[ ] Replace each match with the appropriate
text (and escaped URL markers)
[ ] When all matches are done, find and
replace all URL markers.
*/
func replace_markup(s string, postinfo PostInfo) string {
    type from_to struct {
        from    string
        to      func(s string, m []int) string
    }

    type save_urls struct {
        top int
        urls []string
    }

    saved_urls := save_urls{}

    /*
    [=] Save a URL and return a temporary $$$$<num>$$$$
    url to be used until it is replaced
    */
    save_url := func(save *save_urls, url string) string {
        url = html.UnescapeString(url)
        save.top += 1
        save.urls = append(save.urls, url)
        return `$$$$` + strconv.Itoa(save.top-1) + `$$$$`
    }

    link_replacer := func(s string, m []int) string {
        tmp_url := save_url(&saved_urls, s[m[2]:m[3]])
        path := s[m[4]:m[5]]
        return `<a href="` + tmp_url + `">` + path + `</a>`
    }

    youtube_replacer := func(s string, m []int) string {
        tmp_url := save_url(&saved_urls, s[m[2]:m[3]])
        return `<iframe class=vid src="https://www.youtube.com/embed/` + tmp_url + `" frameborder="0" allowfullscreen></iframe>`
    }

    /*[!] We need to copy the images in each repository
    to the current directory.
    [+] Show a copy message so this can be done manually
    TODO: automate this
    */
    pic_replacer := func(s string, m []int) string {
        url := html.UnescapeString(s[m[2]:m[3]])
        imgsrc := filepath.Join(filepath.Dir(postinfo.InPath), url)
        fmt.Println("cp '" + imgsrc + "' .")
        alt := fname_to_title(url)
        tmp_url := save_url(&saved_urls, url)
        return `<img class=pic src="` + tmp_url + `" alt="` + template.HTMLEscapeString(alt) + `"></img>`
    }

    pic_link_replacer := func(s string, m[]int) string {
        tmp_url := save_url(&saved_urls, s[m[4]:m[5]])
        img := pic_replacer(s, m)
        return `<a href="` + tmp_url + `">` + img + `</a>`
    }

    bold_replacer := func(s string, m []int) string {
        return s[m[2]:m[3]] + `<b>` + s[m[4]:m[5]] + `</b>`
    }

    italic_replacer := func(s string, m []int) string {
        return s[m[2]:m[3]] + `<i>` + s[m[4]:m[5]] + `</i>`
    }

    class_replacer := func(s string, m []int) string {
        n := m[5] - m[4]
        classname := "class" + strconv.Itoa(n)
        return s[m[2]:m[3]] + `<span class=` + classname + `>` + s[m[6]:m[7]] + `</span>`
    }

    ft_maps := []from_to {
        {from: `\[href=([^]]+)\]\(([^)]+)\)`,
           to: link_replacer },
        {from: LINE_MARKER + WHITESPACE + `*([^ ]*.png)` + WHITESPACE + `*\[href=([^]]+)\]`,
           to: pic_link_replacer },
        {from: LINE_MARKER + WHITESPACE + `*https://www.youtube.com/watch\?v=([^ \t\n\r]*)` + WHITESPACE + `*`,
           to: youtube_replacer },
        {from: LINE_MARKER + WHITESPACE + `*([^ ]*.png)` + WHITESPACE + `*`,
           to: pic_replacer },
        {from: `([ \t\n\r])\*([A-Za-z0-9].*?)\*`,
           to: bold_replacer },
        {from: `([ \t\n\r])_([A-Za-z0-9].*?)_`,
           to: italic_replacer },
        {from: `([ \t\n\r])\*([_]+)([A-Za-z0-9].*?)[_]+\*`,
           to: class_replacer },
    }

    for _,ft_map := range ft_maps {
        rx := regexp.MustCompile(ft_map.from)
        m := rx.FindStringSubmatchIndex(s)
        r := ""
        for m != nil {
            r += s[:m[0]] + ft_map.to(s, m)
            s = s[m[1]:]
            m = rx.FindStringSubmatchIndex(s)
        }
        s = r + s
    }

    replace_tmp_urls := func(s string, save save_urls) string {
        rx := regexp.MustCompile(`\$\$\$\$([0-9]+)\$\$\$\$`)
        m := rx.FindStringSubmatchIndex(s)
        r := ""
        for m != nil {
            ndx,err := strconv.Atoi(s[m[2]:m[3]])
            if err != nil || ndx >= len(save.urls) {
                r += s[:m[1]]
                s = s[m[1]:]
            } else {
                r += s[:m[0]] + save.urls[ndx]
                s = s[m[1]:]
            }
            m = rx.FindStringSubmatchIndex(s)
        }
        s = r + s
        return s
    }

    return replace_tmp_urls(s, saved_urls)
}

/*
[=] Return the comment markers
for the type of file passed in.
TODO: Take inputs from external
configuration file.
*/
func get_comment_marker(postinfo PostInfo) (post_comment_marker,error) {
    m := map[string]post_comment_marker {
        ".go": { start: "/**", decorate: '*', end: "*/" },
    }

    ext := filepath.Ext(postinfo.InPath)
    markers,ok := m[ext]
    if !ok {
        return markers, errors.New("Did not find post comment marker for filetype: " + ext)
    }
    return markers,nil
}

/*
[=] Generate the list of blogs
in a new index.html file.
*/
const INDEX_TPL=`<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
    <title>The Productive Programmer</title>
    <meta name="description" content="The blog for programmers who are excited about being productive and want to make the best use of their time">

    <!-- improve view in mobile -->
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
@-ms-viewport{
    width: device-width;
    initial-scale: 1;
}
    </style>

    <!-- favicons -->
    <link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png">
    <link rel="icon" type="image/png" href="/favicon-32x32.png" sizes="32x32">
    <link rel="icon" type="image/png" href="/favicon-16x16.png" sizes="16x16">
    <link rel="manifest" href="/manifest.json">
    <link rel="mask-icon" href="/safari-pinned-tab.svg" color="#5bbad5">
    <meta name="theme-color" content="#ffffff">


    <!-- styling reset -->
    <style>
* { margin: 0; padding: 0; font-family: monospace; font-size: 12px; }
    </style>

    <!-- style -->
    <style>
.main-content { max-width: 640px; }
@media (min-width: 768px) { * { font-size: 14px; } }
@media (min-width: 768px) { .main-content { margin-left: 33vw; } }
.main-content { margin-top: 0; }
.home { margin-bottom: 3em; }
.home img { max-width: 64px; text-align: right; }
.date { margin: 0; }
.toptitle { margin: 5px 0; }
.title { font-weight: bold; margin: 1.67em 0 0.67em 0; }
.file { margin: 0.67em 0 3em 0; }
.content { white-space: pre-wrap; }
.code { white-space: pre; font-size: 75%; color: #999; }
.sep { white-space: pre; }
.mycomment input { font: serif; font-size:95%; display: block; }
.mycomment div { margin: 5px 0; }
.comment { max-width: 240px; }
.comment * { font-family: serif; max-width: 240px; }
.comment div { margin: 5px 0; }
.comment .author { font-weight: bold; white-space: pre-wrap; }
.post { display: block; margin: 0.5em 0; }
@media (max-width: 767px) {
.date,.toptitle,.title,.post,.home,.file,.content,.code,.mycomment,.comments { margin-left: 8px; margin-right: 8px; }
}
    </style>

    <script src='https://www.google.com/recaptcha/api.js'></script>

</head>
<body>
    <div class=main-content>

        <div class=toptitle>The Productive Programmer's Blog</div>

        <div class=home>
            <a href=/><img src=prodprog-bw.png alt='logo'></img></a>
        </div>

        <div class=title>Posts</div>
        {{range .}}
        <span class=post>+ <a href={{urlquery .OutPath}}>{{.HTMLTitle}}</a></span>
        {{end}}

        <div class=sep>
  .  .  .  .  .  .  .  .  .  .  
        </div>

    </div>

</body>
</html>`
func generate_blog_index(pi []PostInfo) error {
    t,err := template.New("index.html").Parse(INDEX_TPL)
    if err != nil {
        return err
    }

    i,err := os.Create("index.html")
    if err != nil {
        return err
    }
    defer i.Close()

    return t.Execute(i, pi)
}

/*
[=] Generate all blog posts
*/
func generate_blog_posts(pi []PostInfo) {
    for _,postinfo := range pi {
        generate_blog_post(postinfo)
    }
}

/*
[=] Generate a blog post
*/
const POST_TPL=`<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
    <title>The Productive Programmer</title>
    <meta name="description" content="The blog for programmers who are excited about being productive and want to make the best use of their time">

    <!-- improve view in mobile -->
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
@-ms-viewport{
    width: device-width;
    initial-scale: 1;
}
    </style>

    <!-- favicons -->
    <link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png">
    <link rel="icon" type="image/png" href="/favicon-32x32.png" sizes="32x32">
    <link rel="icon" type="image/png" href="/favicon-16x16.png" sizes="16x16">
    <link rel="manifest" href="/manifest.json">
    <link rel="mask-icon" href="/safari-pinned-tab.svg" color="#5bbad5">
    <meta name="theme-color" content="#ffffff">


    <!-- styling reset -->
    <style>
* { margin: 0; padding: 0; font-family: monospace; font-size: 12px; }
    </style>

    <!-- style -->
    <style>
.main-content { max-width: 640px; }
@media (min-width: 768px) { * { font-size: 14px; } }
@media (min-width: 768px) { .main-content { margin-left: 33vw; } }
div { margin: 3em 0; }
.main-content { margin-top: 0; }
.home { margin: 0 10%; float: right; }
.home img { max-width: 64px; }
.date { margin: 0; }
.title { font-weight: bold; margin: 0.67em 0; }
.file { margin: 0.67em 0 3em 0; }
.content { white-space: pre-wrap; }
.code { white-space: pre; font-size: 75%; color: #999; }
.sep { white-space: pre; }
.mycomment input { font: serif; font-size:95%; display: block; }
.mycomment div { margin: 5px 0; }
.comment { max-width: 240px; }
.comment * { font-family: serif; max-width: 240px; }
.comment div { margin: 5px 0; }
.comment .author { font-weight: bold; white-space: pre-wrap; }
@media (max-width: 767px) {
.date,.title,.file,.content,.code,.mycomment,.comments { margin-left: 8px; margin-right: 8px; }
}
#submit_comment { font-size: 1.2em; }
    </style>
    <style>
.class1 { color: red; }
.class2 { color: green; }
.class3 { color: blue; }
    </style>

    <script src='https://www.google.com/recaptcha/api.js'></script>

</head>
<body>
    <div class=home>
        <a href=/><img src=prodprog-bw.png alt='logo'></img></a>
    </div>

    <div class=main-content>

        <div class=date>{{html (post_date .)}}</div>

        <div class=title>{{.HTMLTitle}}/</div>

        <div class=file>
            <a href={{gitlab_link .InPath}}>{{html (post_fname .)}}</a>
        </div>

        {{range .Content}}
<div class={{contenttype_class .}}>{{.HTMLVal}}</div>
        {{end}}

        <div class=sep>
  .  .  .  .  .  .  .  .  .  .  
        </div>

        <script>
function enable_submit() {
    document.getElementById('submit_comment').disabled = false;
}
        </script>
        <form class=mycomment method=POST>
            <input type=hidden name=comment_on value="{{urlquery .OutPath}}">
            <textarea name=comment cols=24 rows=8></textarea><br/>
            <input type=text placeholder="Email(optional)" name=email id=email>
            <div class="g-recaptcha" data-callback="enable_submit" data-sitekey="6LeVxgsUAAAAACGmyQCmlk4KJDHn8oCcQvSG-45H"></div>
            <input id=submit_comment disabled=disabled type=submit value="Submit My Comment">
        </form>

        <div class=sep>
  .  .  .  .  .  .  .  .  .  .  
        </div>

<?php
$root = $_SERVER['DOCUMENT_ROOT'];
$config = parse_ini_file($root . '/../php-mysql-config.ini');
$conn = mysqli_connect('localhost', $config['username'], $config['password'], $config['dbname']);
if(! $conn ) {
    die('Could not connect: ' . mysqli_connect_error());
}

if (isset($_POST['comment']) && !empty($_POST['comment'])) {

    if(isset($_POST['g-recaptcha-response']) && !empty($_POST['g-recaptcha-response'])) {

        $secret = "6LeVxgsUAAAAAI5E4n84Lp60ojgBPw0U7AX6aIZV";
        $recaptcha = $_POST['g-recaptcha-response'];

        $url = 'https://www.google.com/recaptcha/api/siteverify';
        $data = 'secret=' . $secret . '&response=' . $recaptcha;

        $ch = curl_init( $url );
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);
        curl_setopt( $ch, CURLOPT_POST, 1);
        curl_setopt( $ch, CURLOPT_POSTFIELDS, $data);
        curl_setopt( $ch, CURLOPT_FOLLOWLOCATION, 1);
        curl_setopt( $ch, CURLOPT_HEADER, 0);
        curl_setopt( $ch, CURLOPT_RETURNTRANSFER, 1);

        $verifyResponse = curl_exec( $ch );

        $responseData = json_decode($verifyResponse);
        if ($responseData->success) {

            $comment_on = mysqli_real_escape_string($conn, $_POST['comment_on']);
            $comment = mysqli_real_escape_string($conn, $_POST['comment']);
            $email = mysqli_real_escape_string($conn, $_POST['email']);

            $addr   = mysqli_real_escape_string($conn, $_SERVER['REMOTE_ADDR']);
            $port   = mysqli_real_escape_string($conn, $_SERVER['REMOTE_PORT']);
            $method = mysqli_real_escape_string($conn, $_SERVER['REQUEST_METHOD']);
            $url    = mysqli_real_escape_string($conn, $_SERVER['REQUEST_URI']);

            $client_ip       = isset($_SERVER['HTTP_CLIENT_IP'])       ? mysqli_real_escape_string($conn, $_SERVER['HTTP_CLIENT_IP']) : '';
            $x_forwarded_for = isset($_SERVER['HTTP_X_FORWARDED_FOR']) ? mysqli_real_escape_string($conn, $_SERVER['HTTP_X_FORWARDED_FOR']) : '';
            $ua              = isset($_SERVER['HTTP_USER_AGENT'])      ? mysqli_real_escape_string($conn, $_SERVER['HTTP_USER_AGENT']) : '';
            $referer         = isset($_SERVER['HTTP_REFERER'])         ? mysqli_real_escape_string($conn, $_SERVER['HTTP_REFERER']) : '';
            $sz              = isset($_SERVER['CONTENT_LENGTH'])       ? mysqli_real_escape_string($conn, $_SERVER['CONTENT_LENGTH']) : '';

            $sql = "insert into comments (comment_on,comment,email,at,addr,client_ip,x_forwarded_for,port,ua,referer) VALUES('$comment_on','$comment','$email',NOW(),'$addr','$client_ip','$x_forwarded_for','$port','$ua','$referer')";

            $retval = mysqli_query($conn, $sql);
            if (!$retval) {
                error_log(mysqli_error($conn));
                mysqli_close($conn);
                die("Uh...oh! Something went wrong!");
            }

        }
    }

}

$sql = "select * from comments where comment_on='{{urlquery .OutPath}}' order by 'at' desc";
$result = mysqli_query($conn, $sql);
if(mysqli_num_rows($result) > 0) {
?>
        <div class=comments>
<?php
    while($row = mysqli_fetch_assoc($result)) {
        $email = htmlspecialchars($row['email']);
        if (!empty($email) && strpos($email, '@')) {
            $author = substr($email, 0, strpos($email, '@'));
        } else {
            $author = "Someone";
        }
        $comment = htmlspecialchars($row['comment']);
        echo "<div class=comment>";
        echo "<div><span class=author>" . $author . "</span> says:</div>";
        echo "<div class=txt>" . $comment . "</div>";
        echo "</div>";
    }
?>
        </div>

        <div class=sep>
  .  .  .  .  .  .  .  .  .  .  
        </div>
<?php

}

mysqli_close($conn);
?>



    </div>

</body>
</html>`
func generate_blog_post(postinfo PostInfo) error {

    var fm = template.FuncMap {
        "post_date" : post_date,
        "post_fname" : post_fname,
        "contenttype_class" : contenttype_class,
        "gitlab_link" : gitlab_link,
    }

    t,err := template.New("post.html").Funcs(fm).Parse(POST_TPL)
    if err != nil {
        return err
    }

    post,err := os.Create(postinfo.OutPath)
    if err != nil {
        return err
    }
    defer post.Close()

    return t.Execute(post, postinfo)
}

func post_date(postinfo PostInfo) string {
    return postinfo.On.Format("Jan 01")
}
func post_fname(postinfo PostInfo) string {
    return filepath.Base(postinfo.OutPath)
}

func contenttype_class(pc PostContent) string {
    if pc.Typ == POSTCOMMENT {
        return "content"
    } else if pc.Typ == CODE {
        return "code"
    } else {
        return "empty"
    }
}

var GITLAB_PFX = "https://gitlab.com/productiveprogrammer/"
func gitlab_link(path string) string {
    paths := strings.Split(path, string(filepath.Separator))
    return GITLAB_PFX + paths[len(paths)-2] + "/blob/master/" + paths[len(paths)-1]
}
