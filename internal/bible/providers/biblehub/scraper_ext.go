package biblehub

// SetBaseURL sets the base URL for the scraper. This is useful for testing.
func (s *Scraper) SetBaseURL(url string) {
	s.baseURL = url
}
