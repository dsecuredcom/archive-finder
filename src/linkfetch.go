package src

func foldersFetched(host string) (folders []string) {

	// 1) Request the host using GET and https
	// 2) if the response is 200, parse the HTML and extract all hrefs
	// 3) if the response is 30x, check if it stays in scope and fetch the new location
	// 4) if the new location is 200, parse the HTML and extract all paths which contain files like images, css, styles, js, etc.
	// 5) if the new location is 30x, stop, return empty list

	// check if the relevant hrefs are internal links, we need to drop all irrelevant ones, we only want internal links of this host
	// this means also dropping all external links, all links that point to a different host

	// what we DO NOT WANT is a list of e.g. links to posts/articles
	//return []string{"folder1", "folder2", "folder3"} without trailing or leading slashes

	// Also please request using sane user agents and timeouts
	// When a folder is overly long, it should be skipped

	// Still thinking here, since its a resource waster
	return []string{}
}
