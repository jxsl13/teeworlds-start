# Scripts

Directory containing backup scripts for the Crontab that periodically executes those tasks.

````
crontab -e
```

IN your case this path migt differ, so change the path to your specifid locations.
Backup SQLite database every 12 hours.
Compress demo files every 30 days.
```shell
0 */12 * * * bash ~/servers/scripts/backup_ranking_db.sh
0 3 * * 1 bash ~/servers/scripts/compress_demos.sh
```