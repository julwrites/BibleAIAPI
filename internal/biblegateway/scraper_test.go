package biblegateway

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

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
			expected: `<h3>For God So Loved the World</h3><p><span class="text John-3-16"><sup class="versenum">16 </sup>“For God so loved the world, that he gave his only Son, that whoever believes in him should not perish but have eternal life. </span></p>`,
		},
		{
			name:     "Matthew 28:19-20",
			book:     "Matthew",
			chapter:  "28",
			verse:    "19-20",
			version:  "ESV",
			htmlFile: "testdata/get_verse_matthew.html",
			expected: `<p><span class="text Matt-28-19"><sup class="versenum">19 </sup>Go therefore and make disciples of all nations, baptizing them in the name of the Father and of the Son and of the Holy Spirit, </span><span class="text Matt-28-20"><sup class="versenum">20 </sup>teaching them to observe all that I have commanded you. And behold, I am with you always, to the end of the age.” </span></p>`,
		},
		{
			name:     "Proverbs 3:5-6",
			book:     "Proverbs",
			chapter:  "3",
			verse:    "5-6",
			version:  "ESV",
			htmlFile: "testdata/get_verse_proverbs.html",
			expected: `<p><span class="text Prov-3-5"><sup class="versenum">5 </sup>Trust in the Lord with all your heart,</span><br/><span class=""><span class="">    </span><span class="text Prov-3-5">and do not lean on your own understanding.</span></span><br/><span class="text Prov-3-6"><sup class="versenum">6 </sup>In all your ways acknowledge him,</span><br/><span class=""><span class="">    </span><span class="text Prov-3-6">and he will make straight your paths.</span></span></p>`,
		},
		{
			name:     "Psalm 121",
			book:     "Psalm",
			chapter:  "121",
			verse:    "",
			version:  "NIV",
			htmlFile: "testdata/get_verse_psalm_121.html",
			expected: `<h3>My Help Comes from the Lord</h3><h4>A Song of Ascents.</h4><p><span class="text Ps-121-1"><sup class="versenum">1 </sup>I lift up my eyes to the hills.</span><br/><span class=""><span class="">    </span><span class="text Ps-121-1">From where does my help come?</span></span><br/><span class="text Ps-121-2"><sup class="versenum">2 </sup>My help comes from the Lord,</span><br/><span class=""><span class="">    </span><span class="text Ps-121-2">who made heaven and earth.</span></span></p><p class="top-1"> </p><p><span class="text Ps-121-3"><sup class="versenum">3 </sup>He will not let your foot be moved;</span><br/><span class=""><span class="">    </span><span class="text Ps-121-3">he who keeps you will not slumber.</span></span><br/><span class="text Ps-121-4"><sup class="versenum">4 </sup>Behold, he who keeps Israel</span><br/><span class=""><span class="">    </span><span class="text Ps-121-4">will neither slumber nor sleep.</span></span></p><p class="top-1"> </p><p><span class="text Ps-121-5"><sup class="versenum">5 </sup>The Lord is your keeper;</span><br/><span class=""><span class="">    </span><span class="text Ps-121-5">the Lord is your shade on your right hand.</span></span><br/><span class="text Ps-121-6"><sup class="versenum">6 </sup>The sun shall not strike you by day,</span><br/><span class=""><span class="">    </span><span class="text Ps-121-6">nor the moon by night.</span></span></p><p class="top-1"> </p><p><span class="text Ps-121-7"><sup class="versenum">7 </sup>The Lord will keep you from all evil;</span><br/><span class=""><span class="">    </span><span class="text Ps-121-7">he will keep your life.</span></span><br/><span class="text Ps-121-8"><sup class="versenum">8 </sup>The Lord will keep</span><br/><span class=""><span class="">    </span><span class="text Ps-121-8">your going out and your coming in</span></span><br/><span class=""><span class="">    </span><span class="text Ps-121-8">from this time forth and forevermore.</span></span></p>`,
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

			if verse != tc.expected {
				t.Errorf("expected verse to be:\n%s\nbut got:\n%s", tc.expected, verse)
			}
		})
	}
}
