package biblegateway

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"bible-api-service/internal/bible"
)

func normalizeSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func TestGetVerse(t *testing.T) {
	testCases := []struct {
		name       string
		book       string
		chapter    string
		verse      string
		version    string
		htmlFile   string
		expected   string
		expectFail bool
	}{
		{
			name:     "John 3:16",
			book:     "John",
			chapter:  "3",
			verse:    "16",
			version:  "ESV",
			htmlFile: "testdata/get_verse_success.html",
			expected: `<h3>For God So Loved the World</h3> <p> <sup>16 </sup>“For God so loved the world, that he gave his only Son, that whoever believes in him should not perish but have eternal life. </p>`,
		},
		{
			name:     "Matthew 28:19-20",
			book:     "Matthew",
			chapter:  "28",
			verse:    "19-20",
			version:  "ESV",
			htmlFile: "testdata/get_verse_matthew.html",
			expected: `<p> <sup>19 </sup>Go therefore and make disciples of all nations, baptizing them in the name of the Father and of the Son and of the Holy Spirit, <sup>20 </sup>teaching them to observe all that I have commanded you. And behold, I am with you always, to the end of the age.” </p>`,
		},
		{
			name:     "Proverbs 3:5-6",
			book:     "Proverbs",
			chapter:  "3",
			verse:    "5-6",
			version:  "ESV",
			htmlFile: "testdata/get_verse_proverbs.html",
			expected: `<p><sup>5 </sup>Trust in the Lord with all your heart,<br/>and do not lean on your own understanding.</p> <p><sup>6 </sup>In all your ways acknowledge him,<br/>and he will make straight your paths.</p>`,
		},
		{
			name:     "Bug Reproduction (John 3:16 with extras)",
			book:     "John",
			chapter:  "3",
			verse:    "16",
			version:  "ESV",
			htmlFile: "testdata/bug_repro.html",
			expected: `<h3>For God So Loved the World</h3> <p> <sup>16 </sup>“For God so loved the world, that he gave his only Son, that whoever believes in him should not perish but have eternal life. </p>`,
		},
		{
			name:     "Psalm 121",
			book:     "Psalm",
			chapter:  "121",
			verse:    "",
			version:  "NIV",
			htmlFile: "testdata/get_verse_psalm_121.html",
			expected: `<h3>My Help Comes from the Lord</h3> <h4>A Song of Ascents.</h4> <p><sup>1 </sup>I lift up my eyes to the hills.<br/>From where does my help come?</p> <p><sup>2 </sup>My help comes from the Lord,<br/>who made heaven and earth.</p> <br/> <p><sup>3 </sup>He will not let your foot be moved;<br/>he who keeps you will not slumber.</p> <p><sup>4 </sup>Behold, he who keeps Israel<br/>will neither slumber nor sleep.</p> <br/> <p><sup>5 </sup>The Lord is your keeper;<br/>the Lord is your shade on your right hand.</p> <p><sup>6 </sup>The sun shall not strike you by day,<br/>nor the moon by night.</p> <br/> <p><sup>7 </sup>The Lord will keep you from all evil;<br/>he will keep your life.</p> <p><sup>8 </sup>The Lord will keep<br/>your going out and your coming in<br/>from this time forth and forevermore.</p>`,
		},
		{
			name:       "verse not found",
			book:       "Invalid",
			chapter:    "1",
			verse:      "1",
			version:    "ESV",
			htmlFile:   "testdata/get_verse_not_found.html",
			expectFail: true,
		},
		{
			name:     "Bug Reproduction Cross-Chapter (John 1:12-2:4)",
			book:     "John",
			chapter:  "1",
			verse:    "12-2:4",
			version:  "ESV",
			htmlFile: "testdata/bug_repro_cross.html",
			expected: `<p> <sup>12 </sup>But to all who did receive him, who believed in his name, he gave the right to become children of God, <sup>13 </sup>who were born, not of blood nor of the will of the flesh nor of the will of man, but of God.</p> <p><sup>14 </sup>And the Word became flesh and dwelt among us, and we have seen his glory, glory as of the only Son from the Father, full of grace and truth. <sup>15 </sup>(John bore witness about him, and cried out, “This was he of whom I said, ‘He who comes after me ranks before me, because he was before me.’”) <sup>16 </sup>For from his fullness we have all received, grace upon grace. <sup>17 </sup>For the law was given through Moses; grace and truth came through Jesus Christ. <sup>18 </sup>No one has ever seen God; God the only Son, who is at the Father&#39;s side, he has made him known.</p> <h3>The Testimony of John the Baptist</h3><p><sup>19 </sup>And this is the testimony of John, when the Jews sent priests and Levites from Jerusalem to ask him, “Who are you?” <sup>20 </sup>He confessed, and did not deny, but confessed, “I am not the Christ.” <sup>21 </sup>And they asked him, “What then? Are you Elijah?” He said, “I am not.” “Are you the Prophet?” And he answered, “No.” <sup>22 </sup>So they said to him, “Who are you? We need to give an answer to those who sent us. What do you say about yourself?” <sup>23 </sup>He said, “I am the voice of one crying out in the wilderness, ‘Make straight the way of the Lord,’ as the prophet Isaiah said.”</p> <p><sup>24 </sup>(Now they had been sent from the Pharisees.) <sup>25 </sup>They asked him, “Then why are you baptizing, if you are neither the Christ, nor Elijah, nor the Prophet?” <sup>26 </sup>John answered them, “I baptize with water, but among you stands one you do not know, <sup>27 </sup>even he who comes after me, the strap of whose sandal I am not worthy to untie.” <sup>28 </sup>These things took place in Bethany across the Jordan, where John was baptizing.</p> <h3>Behold, the Lamb of God</h3><p><sup>29 </sup>The next day he saw Jesus coming toward him, and said, “Behold, the Lamb of God, who takes away the sin of the world! <sup>30 </sup>This is he of whom I said, ‘After me comes a man who ranks before me, because he was before me.’ <sup>31 </sup>I myself did not know him, but for this purpose I came baptizing with water, that he might be revealed to Israel.” <sup>32 </sup>And John bore witness: “I saw the Spirit descend from heaven like a dove, and it remained on him. <sup>33 </sup>I myself did not know him, but he who sent me to baptize with water said to me, ‘He on whom you see the Spirit descend and remain, this is he who baptizes with the Holy Spirit.’ <sup>34 </sup>And I have seen and have borne witness that this is the Son of God.”</p> <h3>Jesus Calls the First Disciples</h3><p><sup>35 </sup>The next day again John was standing with two of his disciples, <sup>36 </sup>and he looked at Jesus as he walked by and said, “Behold, the Lamb of God!” <sup>37 </sup>The two disciples heard him say this, and they followed Jesus. <sup>38 </sup>Jesus turned and saw them following and said to them, “What are you seeking?” And they said to him, “Rabbi” (which means Teacher), “where are you staying?” <sup>39 </sup>He said to them, “Come and you will see.” So they came and saw where he was staying, and they stayed with him that day, for it was about the tenth hour. <sup>40 </sup>One of the two who heard John speak and followed Jesus was Andrew, Simon Peter&#39;s brother. <sup>41 </sup>He first found his own brother Simon and said to him, “We have found the Messiah” (which means Christ). <sup>42 </sup>He brought him to Jesus. Jesus looked at him and said, “You are Simon the son of John. You shall be called Cephas” (which means Peter).</p> <h3>Jesus Calls Philip and Nathanael</h3><p><sup>43 </sup>The next day Jesus decided to go to Galilee. He found Philip and said to him, “Follow me.” <sup>44 </sup>Now Philip was from Bethsaida, the city of Andrew and Peter. <sup>45 </sup>Philip found Nathanael and said to him, “We have found him of whom Moses in the Law and also the prophets wrote, Jesus of Nazareth, the son of Joseph.” <sup>46 </sup>Nathanael said to him, “Can anything good come out of Nazareth?” Philip said to him, “Come and see.” <sup>47 </sup>Jesus saw Nathanael coming toward him and said of him, “Behold, an Israelite indeed, in whom there is no deceit!” <sup>48 </sup>Nathanael said to him, “How do you know me?” Jesus answered him, “Before Philip called you, when you were under the fig tree, I saw you.” <sup>49 </sup>Nathanael answered him, “Rabbi, you are the Son of God! You are the King of Israel!” <sup>50 </sup>Jesus answered him, “Because I said to you, ‘I saw you under the fig tree,’ do you believe? You will see greater things than these.” <sup>51 </sup>And he said to him, “Truly, truly, I say to you, you will see heaven opened, and the angels of God ascending and descending on the Son of Man.”</p> <h3>The Wedding at Cana</h3><p>On the third day there was a wedding at Cana in Galilee, and the mother of Jesus was there. <sup>2 </sup>Jesus also was invited to the wedding with his disciples. <sup>3 </sup>When the wine ran out, the mother of Jesus said to him, “They have no wine.” <sup>4 </sup>And Jesus said to her, “Woman, what does this have to do with me? My hour has not yet come.” </p>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				html, err := os.ReadFile(tc.htmlFile)
				if err != nil {
					t.Fatalf("failed to read mock html file: %v", err)
				}
				fmt.Fprintln(w, string(html))
			}))
			defer server.Close()

			scraper := &Scraper{client: server.Client(), baseURL: server.URL}

			verse, err := scraper.GetVerse(tc.book, tc.chapter, tc.verse, tc.version)
			if tc.expectFail {
				if err == nil {
					t.Fatal("expected an error, but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if normalizeSpace(verse) != normalizeSpace(tc.expected) {
				t.Errorf("expected verse to be:\n%s\nbut got:\n%s", tc.expected, verse)
			}
		})
	}
}

func TestSearchWords(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html, err := os.ReadFile("testdata/search_words_success.html")
		if err != nil {
			t.Fatalf("failed to read mock html file: %v", err)
		}
		fmt.Fprintln(w, string(html))
	}))
	defer server.Close()

	scraper := &Scraper{client: server.Client(), baseURL: server.URL}
	results, err := scraper.SearchWords("love", "ESV")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []bible.SearchResult{
		{
			Verse: "John 3:16",
			Text:  "For God so loved the world, that he gave his only Son, that whoever believes in him should not perish but have eternal life.",
			URL:   server.URL + "/passage/?search=John%203%3A16&version=ESV",
		},
		{
			Verse: "1 John 4:8",
			Text:  "Anyone who does not love does not know God, because God is love.",
			URL:   server.URL + "/passage/?search=1%20John%204%3A8&version=ESV",
		},
	}

	if len(results) != len(expected) {
		t.Fatalf("expected %d results, but got %d", len(expected), len(results))
	}

	for i, result := range results {
		if result.Verse != expected[i].Verse {
			t.Errorf("expected verse to be %s, but got %s", expected[i].Verse, result.Verse)
		}
		if result.Text != expected[i].Text {
			t.Errorf("expected text to be %s, but got %s", expected[i].Text, result.Text)
		}
		if result.URL != expected[i].URL {
			t.Errorf("expected url to be %s, but got %s", expected[i].URL, result.URL)
		}
	}
}

func TestNewScraper(t *testing.T) {
	scraper := NewScraper()
	if scraper.client == nil {
		t.Error("expected client to be initialized")
	}
	if scraper.baseURL != "https://classic.biblegateway.com" {
		t.Errorf("expected baseURL to be https://classic.biblegateway.com, but got %s", scraper.baseURL)
	}
}

func TestGetVerse_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	scraper := &Scraper{client: server.Client(), baseURL: server.URL}
	_, err := scraper.GetVerse("John", "3", "16", "ESV")
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}
}

func TestSearchWords_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	scraper := &Scraper{client: server.Client(), baseURL: server.URL}
	_, err := scraper.SearchWords("love", "ESV")
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}
}

func TestGetVerse_CrossChapterQuery(t *testing.T) {
	// We want to verify that the URL is constructed correctly for cross-chapter references.
	// We don't care about the HTML response here, just the request URL.

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the query string
		q := r.URL.Query()
		search := q.Get("search")

		// Expected: John 1:12-2:4
		// The browser/client might encode it differently, but it should be equivalent.
		// "John 1:12-2:4"

		if search != "John 1:12-2:4" {
			t.Errorf("expected search query 'John 1:12-2:4', got '%s'", search)
		}

		// Return dummy HTML to avoid error
		fmt.Fprintln(w, `<div class="passage-text">OK</div>`)
	}))
	defer server.Close()

	scraper := &Scraper{client: server.Client(), baseURL: server.URL}

	// effectively what happens after ParseVerseReference("John 1:12-2:4")
	// book="John", chapter="1", verse="12-2:4"
	_, err := scraper.GetVerse("John", "1", "12-2:4", "ESV")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
