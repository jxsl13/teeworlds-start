# Teeworlds server configuration files

Naming convention for config files:

```shell
autoexec_zcatch_srv_grenade-01.cfg
```
- File must start with `autoexec_`
- File must contain the name of the server executable `autoexec_zcatch_srv` where `zcatch_srv` is the executable found in `executables`
- File must not use an `_` after `autoexec_zcatch_srv_`
- File can define differentiating suffixes like `autoexec_zcatch_srv_grenade-01`
- File must end in `.cfg`
- File must not be inside of a sub directory of the `configs` directory

## Example

Location: `configs/autoexec_zcatch_srv_grenade-01.cfg`

```bash
sv_name Simply zCatch Grenade

sv_map "ctf5"
sv_port 8303
ec_port 9303

sv_weapon_mode 3
sv_skill_level 1

logfile "logs/GRENADE-01-"
sv_auto_demo_suffix "_GRENADE-01"

exec "configs/shared-zcatch.cfg"
```

Location: `configs/shared-zcatch.cfg`

```bash
sv_motd "shared message of the day"

# ip stuff
sv_max_clients_per_ip 2

# MOderator RCON (F2) password
sv_rcon_mod_password "moderator password, leave empty of not needed or comment out with a # prefix"

# RCON
sv_rcon_password "shared admin password"

# ECON (Telnet connection to the server)
ec_password "shared econ password, leave empty if not needed"
ec_bantime 0
ec_auth_timeout 60
ec_output_level 2

# DEMO
sv_auto_demo_record	1
sv_auto_demo_max 0

# DATABASE
sv_db_type "sqlite"
sv_db_sqlite_file "ranking/ranking.db"

# MODERATOR

# can logout
mod_command logout 1

# can see accessible commands of moderators
mod_command mod_status 1

# can see player data
mod_command status 1

# can force votes with vote yes or vote no
mod_command vote 1

# can send server messages
mod_command say 1

# can move players to spectators if they are afk
mod_command set_team 1

# can see muted players
mod_command mutes 1

# can mute a player
mod_command mute 1

# can unmute a player from the muted list
mod_command unmute 1

# can see vote banned players
mod_command votebans 1

# can voteban a player
mod_command voteban 1

# can unvote ban a player from voteban list
mod_command unvoteban 1

# can unvoteban by client ID
mod_command unvoteban_client 1

# can ban people
mod_command ban 1

# can unban people
mod_command unban 1

# can see bans
mod_command bans 1

# punish cheater
mod_command punish 1

# pardon cheater
mod_command unpunish 1

# view punished cheaters
mod_command punishments 1


add_vote "========== Maps ==========" "say Maps"
add_vote "ctf1" "change_map ctf1"
add_vote "ctf2" "change_map ctf2"
add_vote "ctf3" "change_map ctf3"
add_vote "ctf4" "change_map ctf4"
```

