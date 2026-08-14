package cli

import (
	"blueprint/appinfo"
	"blueprint/models"
	"fmt"
)

func Completion(input models.CommandInput) {
	if len(input.Arguments) == 0 || input.Arguments[0] != "bash" {
		fmt.Printf("Usage: %s completion bash\n", appinfo.CLIName)
		return
	}

	fmt.Printf(`_%s_completion() {
    local cur
    local suggestions
    local args=()

    cur="${COMP_WORDS[COMP_CWORD]}"

    if (( COMP_CWORD > 1 )); then
        args=("${COMP_WORDS[@]:1:$((COMP_CWORD-1))}")
    fi

    args+=("$cur")
    suggestions=$(%s __complete "${args[@]}")

    COMPREPLY=($(compgen -W "$suggestions" -- "$cur"))
}

complete -F _%s_completion %s
`, appinfo.CLIName, appinfo.CLIName, appinfo.CLIName, appinfo.CLIName)
}
