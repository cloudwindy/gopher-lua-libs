// Package plugin implements run lua-code from lua-code.
package plugin

import (
	"context"
	"sync"

	filepath "github.com/vadv/gopher-lua-libs/filepath"
	http "github.com/vadv/gopher-lua-libs/http"
	inspect "github.com/vadv/gopher-lua-libs/inspect"
	ioutil "github.com/vadv/gopher-lua-libs/ioutil"
	json "github.com/vadv/gopher-lua-libs/json"
	regexp "github.com/vadv/gopher-lua-libs/regexp"
	strings "github.com/vadv/gopher-lua-libs/strings"
	tac "github.com/vadv/gopher-lua-libs/tac"
	tcp "github.com/vadv/gopher-lua-libs/tcp"
	time "github.com/vadv/gopher-lua-libs/time"
	xmlpath "github.com/vadv/gopher-lua-libs/xmlpath"
	yaml "github.com/vadv/gopher-lua-libs/yaml"

	lua "github.com/yuin/gopher-lua"
)

type luaPlugin struct {
	sync.Mutex
	state      *lua.LState
	cancelFunc context.CancelFunc
	running    bool
	error      error
	body       string
}

func (p *luaPlugin) getError() error {
	p.Lock()
	defer p.Unlock()
	err := p.error
	return err
}

func (p *luaPlugin) getRunning() bool {
	p.Lock()
	defer p.Unlock()
	running := p.running
	return running
}

func (p *luaPlugin) setError(err error) {
	p.Lock()
	defer p.Unlock()
	p.error = err
}

func (p *luaPlugin) setRunning(val bool) {
	p.Lock()
	defer p.Unlock()
	p.running = val
}

func (p *luaPlugin) start() {
	p.Lock()
	state := lua.NewState()
	// preload all
	filepath.Preload(state)
	http.Preload(state)
	inspect.Preload(state)
	ioutil.Preload(state)
	json.Preload(state)
	regexp.Preload(state)
	strings.Preload(state)
	tac.Preload(state)
	tcp.Preload(state)
	time.Preload(state)
	xmlpath.Preload(state)
	yaml.Preload(state)
	//
	p.state = state
	p.error = nil
	p.running = true
	ctx, cancelFunc := context.WithCancel(context.Background())
	p.state.SetContext(ctx)
	p.cancelFunc = cancelFunc
	p.Unlock()

	// blocking
	p.setError(p.state.DoString(p.body))
	p.setRunning(false)
}

func checkPlugin(L *lua.LState, n int) *luaPlugin {
	ud := L.CheckUserData(n)
	if v, ok := ud.Value.(*luaPlugin); ok {
		return v
	}
	L.ArgError(n, "plugin expected")
	return nil
}

// New(): lua plugin.new(body) returns plugin_ud
func New(L *lua.LState) int {
	body := L.CheckString(1)
	p := &luaPlugin{body: body}
	ud := L.NewUserData()
	ud.Value = p
	L.SetMetatable(ud, L.GetTypeMetatable(`plugin_ud`))
	L.Push(ud)
	return 1
}

// Run(): lua plugin_ud:run()
func Run(L *lua.LState) int {
	p := checkPlugin(L, 1)
	go p.start()
	return 0
}

// IsRunning(): lua plugin_ud:is_running()
func IsRunning(L *lua.LState) int {
	p := checkPlugin(L, 1)
	L.Push(lua.LBool(p.getRunning()))
	return 1
}

// Error(): lua plugin_ud:error() returns err
func Error(L *lua.LState) int {
	p := checkPlugin(L, 1)
	err := p.getError()
	if err == nil {
		return 0
	}
	L.Push(lua.LString(err.Error()))
	return 1
}

// Stop(): lua plugin_ud:stop()
func Stop(L *lua.LState) int {
	p := checkPlugin(L, 1)
	p.Lock()
	defer p.Unlock()
	p.cancelFunc()
	return 0
}
