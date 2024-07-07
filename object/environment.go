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

	if !ok && env.outer != nil {
		obj, ok = env.outer.Get(name)
		//if var not in this env, try the unscoped one :)
		// Outer = "Global", Env="local"
	}
	return obj, ok
}

func (env *Environment) Set(name string, obj Object) Object {
	env.store[name] = obj
	return obj
}

func ScopedEnv(outer *Environment) *Environment {
	//Really Simple solution to scoped functions
	// We create a new Environment (local) for the function,
	// now, that function interacts only with the local scope of variables and function calls
	// however, the function can still go outside and touch grass by using env.outer
	env := NewEnvironment()
	env.outer = outer

	return env
}
