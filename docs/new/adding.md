# Adding a Component


1. Get runcmd and create a constant in constants/kafka/kafka.go (replace kafka) (kuro/tests has helper to gen runcmd)

2. For any direct Packages Available add to CloudInitPkg - (constants/packages.go)

3. Define the Case statements for (configuration/ubuntu.go if Packages are New)

```go
// existing switch... 
	case constants.Clickhouse:
		return db.CLICKHOUSE_RUNCMD

```

4.  Add the Dependency as an enum with the name (constants/constants.go)

```go
func CreateClickhouseUserData(username, pass, vmname, sshpub string) string {
	config, err := configuration.NewConfigBuilder(
		configuration.DefaultPreset{},
		constants.Ubuntu, 
		[]constants.Dependency{
			constants.Zsh,
			constants.PostgresNewComponent, // Add the Component Here 
		},

```
5. Add the CreateComponentUserData function (configuration/presets/component.go)

```go
// add it as a Dependency here
type Dependency string

```


6. Add the cli flag and define func call : `/home/kuro/Documents/Code/Go/kvmgo/cli/flags.go:437`

```go
// Generates the VM according to Presets such as Kubernetes, Spark, Hadoop, and more
func CreateUserdataFromPreset(ctx context.Context, wg *sync.WaitGroup, preset, launch_vm, sshpub string) string {
	log.Print(utils.TurnValBoldColor("Preset: ", preset, utils.PURP_HI))
	switch preset {
	case "kafka":
		return presets.CreateKafkaUserData("ubuntu", "password", launch_vm, sshpub)
	case "hadoop":

```

7. Launch and Test VM

```bash
	go run main.go --launch-vm=ch --preset=clickhouse --mem=8192 --cpu=4

```


```bash

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

go get github.com/jedib0t/go-pretty/v6

go get github.com/jedib0t/go-pretty/v6/text

go get github.com/jedib0t/go-pretty/v6/table

# ubuntu
alias ll="eza -lahs newest --no-permissions --no-user --time-style=+'%b %d %a %H:%M' --total-size --group-directories-first"

alias ll="eza -lahs newest --no-permissions --no-user --time-style=+'%b %d %a %H:%M' --total-size --group-directories-first"

```

