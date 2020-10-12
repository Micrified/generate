package main

import (
	"os"
	"fmt"
	"strings"
	"text/template"
	"io/ioutil"
	"errors"
	"strconv"
	"bufio"
)

// Describes complete ROS application
type Application struct {
	Name     string        // Name of application
	Packages *[]*string    // List of packages to include
	Executors *[]*Executor // Files contained in app
}

// Describes organization of application executables
type Executor struct {
	Name     string        // Name of file
	Type     string        // Class of executor
	IsPrio   bool          // If priority semantics
	Includes *[]*string    // Include directives
	Params   *[]*string    // Constructor parameters
	Nodes    *[]*Node      // Classes contained in file
}

// Describes a shared data container (class)
type Node struct {
	Name     string        // Class names
	Params   *[]*string    // Constructor parameters
	Methods  *[]*Method    // Methods contained in class
}

// Describes an executable unit
type Method struct {
	Name     string        // Method signatures
	MsgType  string        // Type to provide to template
	Params   *[]*string    // Method parameters
	Priority int64         // Priority (if using PPE)
	IsTimer  bool          // If timer
	Period   int64         // Period of timer, if timer
	IsSync   bool          // If should synchronize on subs
	Subs     *[]*string    // Subscribers (ignored if timer)
	Pubs     *[]*string    // Publishers
	WCET     int64         // Nanoseconds
}

// Generate application at specified path
func Generate_Application (app *Application, path string) error {
	var err error = nil

	// check: valid input
	if nil == app {
		return errors.New("bad argument: null pointer")
	}

	// check: valid executors
	if nil == app.Executors {
		return errors.New("app.Executors is null")
	}

	// If path, and ends with slash, strip it 
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	// Layout directory structure
	root_dir := path + "/" + app.Name
	src_dir, include_dir_1 := root_dir + "/src", root_dir + "/include"
	include_dir_2 := include_dir_1 + "/" + app.Name

	// Make directories
	err = os.Mkdir(root_dir, 0777)
	if nil != err {
		return errors.New("Cannot make root directory (" + root_dir + "): " + err.Error())
	}
	err = os.Mkdir(src_dir, 0777)
	if nil != err {
		return errors.New("Cannot make source directory (" + root_dir + "): " + err.Error())
	}
	err = os.Mkdir(include_dir_1, 0777)
	if nil != err {
		return errors.New("Cannot make header directory (" + root_dir + "): " + err.Error())
	}
	err = os.Mkdir(include_dir_2, 0777)
	if err != nil {
		return errors.New("Cannot make header directory (" + root_dir + "): " + err.Error())
	}

	// Generate source files
	for i, executor_p := range *(app.Executors) {
		filepath := src_dir + "/executor_" + strconv.Itoa(i) + ".cpp"
		err = generate_from_template(executor_p, "templates/executor.tmpl", filepath)
		if nil != err {
			return errors.New("Unable to generate executor: " + err.Error())
		}
	}

	// Generate makefile
	err = generate_from_template(app, "templates/CMakeLists.tmpl", root_dir + "/CMakeLists.txt")
	if err != nil {
		return errors.New("Unable to generate CMakeLists: " + err.Error())
	}

	// Generate package descriptor file
	err = generate_from_template(app, "templates/package.tmpl", root_dir + "/package.xml")
	if err != nil {
		return errors.New("Unable to generate package xml: " + err.Error())
	}

	return nil
}

func generate_from_template (data interface{}, in_path, out_path string) error {
	var t *template.Template = nil
	var err error = nil
	var out_file *os.File = nil
	var template_file []byte = []byte{}

	// check: valid input
	if nil == data {
		return errors.New("bad argument: null pointer")
	}
	// Yes, you can use == with string comparisons in go
	if in_path == out_path {
		return errors.New("input file (template) cannot be same as output file")
	}

	// Create the output file
	out_file, err = os.Create(out_path)
	if nil != err {
		return errors.New("unable to create output file (" + out_path + "): " + err.Error())
	}
	defer out_file.Close()

	// Open the template file
	template_file, err = ioutil.ReadFile(in_path)
	if nil != err {
		return errors.New("unable to read input file (" + in_path + "): " + err.Error())
	}
	if template_file == nil {
		panic(errors.New("Nil pointer to read file"))
	}

	t, err = template.New("Unnamed").Parse(string(template_file))
	fmt.Println("parsed!")
	if nil != err {
		return errors.New("unable to parse the template: " + err.Error())
	}

	// Create buffered writer
	writer := bufio.NewWriter(out_file)
	defer writer.Flush()

	// Execute template
	err = t.Execute(writer, data)
	if nil != err {
		return errors.New("error executing template: " + err.Error())
	}

	return nil
}

func main () {

	msg_type, msg_param, topic := "std_msgs::msg::Int64", "std_msgs::msg::Int64::SharedPtr msg_p", "topic_sense"
	subs_1, pubs_1 := &[]*string{}, &[]*string{&topic}
	subs_2, pubs_2 := &[]*string{&topic}, &[]*string{}
	prio_1, prio_2 := int64(0), int64(0)
	name_1, name_2 := "sensor", "on_sensor"
	params_1, params_2 := &[]*string{}, &[]*string{&msg_param}
	is_timer_1, is_timer_2 := true, false
	period_1, period_2 := int64(1000000000), int64(0)
	wcet_1, wcet_2 := int64(1000), int64(100000)
	sync_1, sync_2 := false, false

	n_param := "rclcpp::NodeOptions().start_parameter_event_publisher(false)"
	n_name_1, n_name_2 := "sensor", "controller"
	n_pars_1, n_pars_2 := &[]*string{&n_param}, &[]*string{&n_param}

	// Make methods
	m1 := Method{Name: name_1, MsgType: msg_type, Params: params_1, Priority: prio_1, IsTimer: is_timer_1, Period: period_1, IsSync: sync_1, Subs: subs_1, Pubs: pubs_1, WCET: wcet_1}
	m2 := Method{Name: name_2, MsgType: msg_type, Params: params_2, Priority: prio_2, IsTimer: is_timer_2, Period: period_2, IsSync: sync_2, Subs: subs_2, Pubs: pubs_2, WCET: wcet_2}

	// Make nodes
	n1 := Node{Name: n_name_1, Params: n_pars_1, Methods: &[]*Method{&m1}}
	n2 := Node{Name: n_name_2, Params: n_pars_2, Methods: &[]*Method{&m2}}

	// Make executor
	include := "std_msgs/msg/int64.hpp"
	executor := Executor{Name: "Foo", Type: "rclcpp::executors::SingleThreadedExecutor", IsPrio: false, Includes: &[]*string{&include}, Params: &[]*string{}, Nodes: &[]*Node{&n1, &n2}}

	// Make app
	package_1 := "std_msgs"
	app := Application{Name: "automatic", Packages: &[]*string{&package_1}, Executors: &[]*Executor{&executor}}

		// Build
	err := Generate_Application(&app, "./")
	if nil != err {
		panic(err)
	}
}

