package muxie

import (
	"net/http"
	"strings"
)

const (
	// ParamStart is the character, as a string, which a path pattern starts to define its named parameter.
	ParamStart = ":"
	// WildcardParamStart is the character, as a string, which a path pattern starts to define its named parameter for wildcards.
	// It allows everything else after that path prefix
	// but the Trie checks for static paths and named parameters before that in order to support everything that other implementations do not,
	// and if nothing else found then it tries to find the closest wildcard path(super and unique).
	WildcardParamStart = "*"
)

// Trie contains the main logic for adding and searching nodes for path segments.
// It supports wildcard and named path parameters.
// Trie supports very coblex and useful path patterns for routes.
// The Trie checks for static paths(path without : or *) and named parameters before that in order to support everything that other implementations do not,
// and if nothing else found then it tries to find the closest wildcard path(super and unique).
type Trie struct {
	root *Node

	// if true then it will handle any path if not other parent wildcard exists,
	// so even 404 (on http services) is up to it, see Trie#Insert.
	hasRootWildcard bool

	hasRootSlash bool
}

// NewTrie returns a new, empty Trie.
// It is only useful for end-developers that want to design their own mux/router based on my trie implementation.
//
// See `Trie`
func NewTrie() *Trie {
	return &Trie{
		root:            NewNode(),
		hasRootWildcard: false,
	}
}

// InsertOption is just a function which accepts a pointer to a Node which can alt its `Handler`, `Tag` and `Data`  fields.
//
// See `WithHandler`, `WithTag` and `WithData`.
type InsertOption func(*Node)

// WithHandler sets the node's `Handler` field (useful for HTTP).
func WithHandler(handler http.Handler) InsertOption {
	if handler == nil {
		panic("muxie/WithHandler: empty handler")
	}

	return func(n *Node) {
		if n.Handler == nil {
			n.Handler = handler
		}
	}
}

// WithTag sets the node's `Tag` field (may be useful for HTTP).
func WithTag(tag string) InsertOption {
	return func(n *Node) {
		if n.Tag == "" {
			n.Tag = tag
		}
	}
}

// WithData sets the node's optionally `Data` field.
func WithData(data interface{}) InsertOption {
	return func(n *Node) {
		// data can be replaced.
		n.Data = data
	}
}

// Insert adds a node to the trie.
func (t *Trie) Insert(pattern string, options ...InsertOption) {
	if pattern == "" {
		panic("muxie/trie#Insert: empty pattern")
	}

	n := t.insert(pattern, "", nil, nil)
	for _, opt := range options {
		opt(n)
	}
}

const (
	pathSep  = "/"
	pathSepB = '/'
)

func slowPathSplit(path string) []string {
	if path == pathSep {
		return []string{pathSep}
	}

	// remove last sep if any.
	if path[len(path)-1] == pathSepB {
		path = path[:len(path)-1]
	}

	return strings.Split(path, pathSep)[1:]
}

func resolveStaticPart(key string) string {
	i := strings.Index(key, ParamStart)
	if i == -1 {
		i = strings.Index(key, WildcardParamStart)
	}
	if i == -1 {
		i = len(key)
	}

	return key[:i]
}

func (t *Trie) insert(key, tag string, optionalData interface{}, handler http.Handler) *Node {
	input := slowPathSplit(key)

	n := t.root
	if key == pathSep {
		t.hasRootSlash = true
	}

	var paramKeys []string

	for _, s := range input {
		c := s[0]

		if isParam, isWildcard := c == ParamStart[0], c == WildcardParamStart[0]; isParam || isWildcard {
			n.hasDynamicChild = true
			paramKeys = append(paramKeys, s[1:]) // without : or *.

			// if node has already a wildcard, don't force a value, check for true only.
			if isParam {
				n.childNamedParameter = true
				s = ParamStart
			}

			if isWildcard {
				n.childWildcardParameter = true
				s = WildcardParamStart
				if t.root == n {
					t.hasRootWildcard = true
				}
			}
		}

		if !n.hasChild(s) {
			child := NewNode()
			n.addChild(s, child)
		}

		n = n.getChild(s)
	}

	n.Tag = tag
	n.Handler = handler
	n.Data = optionalData

	n.paramKeys = paramKeys
	n.key = key
	n.staticKey = resolveStaticPart(key)
	n.end = true

	return n
}

// SearchPrefix returns the last node which holds the key which starts with "prefix".
func (t *Trie) SearchPrefix(prefix string) *Node {
	input := slowPathSplit(prefix)
	n := t.root

	for i := 0; i < len(input); i++ {
		s := input[i]
		if child := n.getChild(s); child != nil {
			n = child
			continue
		}

		return nil
	}

	return n
}

// Parents returns the list of nodes that a node with "prefix" key belongs to.
func (t *Trie) Parents(prefix string) (parents []*Node) {
	n := t.SearchPrefix(prefix)
	if n != nil {
		// without this node.
		n = n.Parent()
		for {
			if n == nil {
				break
			}

			if n.IsEnd() {
				parents = append(parents, n)
			}

			n = n.Parent()
		}
	}

	return
}

// HasPrefix returns true if "prefix" is found inside the registered nodes.
func (t *Trie) HasPrefix(prefix string) bool {
	return t.SearchPrefix(prefix) != nil
}

// Autocomplete returns the keys that starts with "prefix",
// this is useful for custom search-engines built on top of my trie implementation.
func (t *Trie) Autocomplete(prefix string, sorter NodeKeysSorter) (list []string) {
	n := t.SearchPrefix(prefix)
	if n != nil {
		list = n.Keys(sorter)
	}
	return
}

// ParamsSetter is the interface which should be implemented by the
// params writer for `Search` in order to store the found named path parameters, if any.
type ParamsSetter interface {
	Set(string, string)
}

// Search is the most important part of the Trie.
// It will try to find the responsible node for a specific query (or a request path for HTTP endpoints).
//
// Search supports searching for static paths(path without : or *) and paths that contain
// named parameters or wildcards.
// Priority as:
// 1. static paths
// 2. named parameters with ":"
// 3. wildcards
// 4. closest wildcard if not found, if any
// 5. root wildcard
func (t *Trie) Search(q string, params ParamsSetter) *Node {
	end := len(q)

	if end == 0 || (end == 1 && q[0] == pathSepB) {
		// fixes only root wildcard but no / registered at.
		if t.hasRootSlash {
			return t.root.getChild(pathSep)
		} else if t.hasRootWildcard {
			// no need to going through setting parameters, this one has not but it is wildcard.
			return t.root.getChild(WildcardParamStart)
		}

		return nil
	}

	n := t.root
	start := 1
	i := 1
	var paramValues []string

	for {
		if i == end || q[i] == pathSepB {
			if child := n.getChild(q[start:i]); child != nil {
				n = child
			} else if n.childNamedParameter { // && n.childWildcardParameter == false {
				n = n.getChild(ParamStart)
				if ln := len(paramValues); cap(paramValues) > ln {
					paramValues = paramValues[:ln+1]
					paramValues[ln] = q[start:i]
				} else {
					paramValues = append(paramValues, q[start:i])
				}
			} else if n.childWildcardParameter {
				n = n.getChild(WildcardParamStart)
				if ln := len(paramValues); cap(paramValues) > ln {
					paramValues = paramValues[:ln+1]
					paramValues[ln] = q[start:]
				} else {
					paramValues = append(paramValues, q[start:])
				}
				break
			} else {
				n = n.findClosestParentWildcardNode()
				if n != nil {
					// means that it has :param/static and *wildcard, we go trhough the :param
					// but the next path segment is not the /static, so go back to *wildcard
					// instead of not found.
					//
					// Fixes:
					// /hello/*p
					// /hello/:p1/static/:p2
					// req: http://localhost:8080/hello/dsadsa/static/dsadsa => found
					// req: http://localhost:8080/hello/dsadsa => but not found!
					// and
					// /second/wild/*p
					// /second/wild/static/otherstatic/
					// req: /second/wild/static/otherstatic/random => but not found!
					params.Set(n.paramKeys[0], q[len(n.staticKey):])
					return n
				}

				return nil
			}

			if i == end {
				break
			}

			i++
			start = i
			continue
		}

		i++
	}

	if n == nil || !n.end {
		if n != nil { // we need it on both places, on last segment (below) or on the first unnknown (above).
			if n = n.findClosestParentWildcardNode(); n != nil {
				params.Set(n.paramKeys[0], q[len(n.staticKey):])
				return n
			}
		}

		if t.hasRootWildcard {
			// that's the case for root wildcard, tests are passing
			// even without it but stick with it for reference.
			// Note ote that something like:
			// Routes: /other2/*myparam and /other2/static
			// Reqs: /other2/staticed will be handled
			// by the /other2/*myparam and not the root wildcard (see above), which is what we want.
			n = t.root.getChild(WildcardParamStart)
			params.Set(n.paramKeys[0], q[1:])
			return n
		}

		return nil
	}

	for i, paramValue := range paramValues {
		if len(n.paramKeys) > i {
			params.Set(n.paramKeys[i], paramValue)
		}
	}

	return n
}
