package object

type Environment struct {
	store map[string]Object
	outer *Environment
}

func NewEnvironment() *Environment {
	return &Environment{store: make(map[string]Object), outer: nil}
}

func (env *Environment) Get(name string) (Object, bool) {
	obj, ok := env.store[name]
	//if not in this env, try the scoped one :)
	// This is really good for any type of functions that deal with scoped parameters
	// Functions and classes(TO DO) have to deal with local vars as well as parameters
	if !ok && env.outer != nil {
		obj, ok = env.outer.Get(name)
	}
	return obj, ok
}

func (env *Environment) Set(name string, obj Object) Object {
	env.store[name] = obj
	return obj
}

func ScopedEnv(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer

	return env
}
