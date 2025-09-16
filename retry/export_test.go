package retry

func IsNonRetriable(err error, nonRetriablePatterns []string) bool {
	return isNonRetriable(err, nonRetriablePatterns)
}
