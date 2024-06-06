package main

import (
	"bufio"
	//	"io/ioutil"
	"os"
	"fmt"
	"strings"
	"regexp"
	"slices"
	//"encoding/json"
)

const (
	INBUFLEN = 65536
)

type node struct{
	Islast bool
	Child map[string]*node
}

var (
	ttl_re = regexp.MustCompile(`^\d+$`)
	label_re = regexp.MustCompile(`^[A-Za-z0-9_-]{1,63}$`)

	unchanged []string // Lines we will pass through from in to out unchanged
	policy_tree = make(map[string]*node)
	output []string // Rules post summarization
)


func walk_tree(nodeptr *node, name string, policy string, ttl string) {

	if nodeptr.Islast {
		output = append(output, fmt.Sprintf("%s IN %s CNAME %s", name, ttl, policy))
	}

	if _, exists := nodeptr.Child["*"]; exists {
		output = append(output, fmt.Sprintf("*.%s IN %s CNAME %s", name, ttl, policy))

	} else {

		for k, v := range nodeptr.Child {
			walk_tree(v, k + "." + name, policy, ttl)
		}
	}
}


func main() {

	input := bufio.NewScanner(os.Stdin)
    scanbuffer := make([]byte, INBUFLEN)
    input.Buffer(scanbuffer, INBUFLEN)

	exit := false

	linenum := 0
	entries := 0;
INPUT:
	for !exit {

		ok := input.Scan()

		if !ok {
			exit = true
			break
		}

		line := input.Text()
		linenum++

        if len(line) == 0 {
            continue INPUT
        }

		tokens := strings.Split(line, " ")

		// name TTL IN CNAME policy
		if len(tokens) != 5 {
			fmt.Fprintf(os.Stderr, "Line %-5d: can not parse \"%s\" as RPZ entry\n", linenum, line)
			unchanged = append(unchanged, line)
			continue INPUT
		}

		name_tok := tokens[0]
		ttl_tok := tokens[1]
		in_tok := tokens[2]
		cname_tok := tokens[3]
		target_tok := tokens[4]

		if in_tok != "IN" {
			fmt.Fprintf(os.Stderr, "Line %-5d: RPZ entry \"%s\" must be in class \"IN\"\n", linenum, line)
			unchanged = append(unchanged, line)
			continue INPUT
		}

		if cname_tok != "CNAME" {
			fmt.Fprintf(os.Stderr, "Line %-5d: RPZ entry \"%s\" must be a CNAME\n", linenum, line)
			unchanged = append(unchanged, line)
			continue INPUT
		}

		// Numeric TTL check
		if !ttl_re.Match([]byte(ttl_tok)) {
			fmt.Fprintf(os.Stderr, "Line %-5d: RPZ entry \"%s\" must have numeric TTL\n", linenum, line)
			unchanged = append(unchanged, line)
			continue INPUT
		}

		// Make the target policy if we haven't seen it yet
		if _, exists := policy_tree[target_tok]; !exists {
			policy_tree[target_tok] = new(node)
			policy_tree[target_tok].Islast = false
			policy_tree[target_tok].Child = make(map[string]*node)
		}

		// Make the TTL if we haven't seen it yet
		if _, exists := policy_tree[target_tok].Child[ttl_tok]; !exists {
			policy_tree[target_tok].Child[ttl_tok] = new(node)
			policy_tree[target_tok].Child[ttl_tok].Islast = false
			policy_tree[target_tok].Child[ttl_tok].Child = make(map[string]*node)
		}


		labels := strings.Split(name_tok, ".")
		slices.Reverse(labels)

		if len(labels) < 2  {
			fmt.Fprintf(os.Stderr, "Line %-5d: RPZ entry \"%s\" has too short a name\n", linenum, line)
			unchanged = append(unchanged, line)
			continue INPUT
		}

		// First label must be blank (name ended with trailing .)
		if labels[0] != "" {
			fmt.Fprintf(os.Stderr, "Line %-5d: RPZ entry \"%s\" has a name without a trailing .\n", linenum, line)
			unchanged = append(unchanged, line)
			continue INPUT
		}

		// Skip first label now
		labels = labels[1:]

		for i, l := range labels {
			if ((i >= len(labels) - 1) && l != "*") && !label_re.Match([]byte(l)) {
				fmt.Fprintf(os.Stderr, "Line %-5d: RPZ entry \"%s\" has unexpected label \"%s\"\n", linenum, line, l)
				unchanged = append(unchanged, line)
				continue INPUT
			}

		}

		// Now do the "hard" work of putting this name into the tree
		end := len(labels) - 1
		nodeptr := policy_tree[target_tok].Child[ttl_tok]
		for i, l := range labels {
			if _, exists := nodeptr.Child[l]; !exists {
				nodeptr.Child[l] = new(node)
				nodeptr.Child[l].Islast = i >= end
				nodeptr.Child[l].Child = make(map[string]*node)
			}

			nodeptr = nodeptr.Child[l]
		}
		entries++

	} // end INPUT

	// Now summarize per policy & ttl
	for p, pn := range policy_tree {
		for t, tn := range pn.Child {
			walk_tree(tn, "", p, t)
		}
	}

	// Send the lines that should not change straight through
	for _, l := range unchanged {
		fmt.Fprintf(os.Stdout, "%s\n", l)
	}

	// Now send just the summarized entries
	for _, l := range output {
		fmt.Fprintf(os.Stdout, "%s\n", l)
	}

	//debug_bytes, _ := json.MarshalIndent(policy_tree, "\t", "\t")
	//fmt.Fprintf(os.Stderr, "%s\n", string(debug_bytes))

	fmt.Fprintf(os.Stderr, "== ZONE STATS ==\n")
	fmt.Fprintf(os.Stderr, "Lines: %d\n", linenum)
	fmt.Fprintf(os.Stderr, "Unparsable: %d\n", len(unchanged))
	fmt.Fprintf(os.Stderr, "Policy targets: %d\n", len(policy_tree))
	fmt.Fprintf(os.Stderr, "Entries in tree: %d\n", entries)
	if entries > 0 {
		fmt.Fprintf(os.Stderr, "Entries after summarization: %d (%4.02f%% reduction)\n", len(output), 100.0 * float64(entries - len(output)) / float64(entries))
	}
}
