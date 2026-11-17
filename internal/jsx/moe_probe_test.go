package jsx

import (
	"context"
	"github.com/MoeForge/moejs"
	"os"
	"testing"
	"time"
)

func TestMoeProbe(t *testing.T) {
	src, e := os.ReadFile("assets/mdx-compiler.js")
	if e != nil {
		t.Fatal(e)
	}
	m, e := moejs.Compile("mdx-compiler.js", string(src))
	if e != nil {
		t.Fatal(e)
	}
	r := moejs.NewRuntime(moejs.Options{MemoryLimit: 256 << 20, DisableDynamicCode: true})
	c, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	stop := context.AfterFunc(c, func() { r.Interrupt(c.Err()) })
	defer stop()
	if e = r.Load(m); e != nil {
		t.Fatal(e)
	}
	h, e := m.Hook("compileMDX")
	if e != nil {
		t.Fatal(e)
	}
	v, e := r.Call(h, moejs.String(`{"source":"# Hello\n\n<Counter initial={2+3} />\n","namespace":"0123456789abcdef0123456789abcdef"}`))
	if e != nil {
		t.Fatal(e)
	}
	var got any
	if e = r.Unmarshal(v, &got); e != nil {
		t.Fatal(e)
	}
	t.Logf("%+v", got)
}
