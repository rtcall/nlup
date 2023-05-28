package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	ns := flag.String("n", "", "nameserver to use")
	v := flag.Bool("v", false, "verbose output")
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Printf("usage: %s [-n nameserver] [-v] name\n", os.Args[0])
		os.Exit(1)
	}

	if *ns == "" {
		if s, err := findNsAddr(); err != nil {
			fmt.Printf("error: %s\n", err)
			os.Exit(1)
		} else {
			*ns = s
		}
	}

	name := flag.Arg(0)
	c, err := dial(*ns)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}

	ans, err := c.sendQuery(name)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}

	if *v {
		fmt.Printf("nameserver: %s\ntype: %d\nttl: %d\n",
			*ns, ans.Ntype, ans.Ttl)
	}

	fmt.Printf("%s: %d.%d.%d.%d\n", name,
		(ans.Addr>>24)&0xff, (ans.Addr>>16)&0xff, (ans.Addr>>8)&0xff,
		ans.Addr&0xff)
}
