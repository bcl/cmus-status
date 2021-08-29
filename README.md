Simple cmus status display

Add this to the tmux status line with:

    set -g status-right "#[fg=green]#(~/bin/cmus-status)#[fg=default] | [#H] %H:%M %e-%b-%g"

Arguments:

* `--volume` shows the volume 0-100%
* `--elapsed` shows the duration and elapsed time like `2m40s/3m43s`
* `--width` sets the total width available, defaults to 60
