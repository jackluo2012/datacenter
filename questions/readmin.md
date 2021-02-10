### 问答抽奖服务

```
goctl rpc template -o rpc/questions.proto
goctl rpc proto -src questions.proto -dir .


goctl model mysql datasource -url="root:admin@tcp(127.0.0.1:3306)/datacenter" -table="app_questions" -dir ./model -c
goctl model mysql datasource -url="root:admin@tcp(127.0.0.1:3306)/datacenter" -table="app_questions_activities" -dir ./model -c
goctl model mysql datasource -url="root:admin@tcp(127.0.0.1:3306)/datacenter" -table="app_questions_answers" -dir ./model -c
goctl model mysql datasource -url="root:admin@tcp(127.0.0.1:3306)/datacenter" -table="app_questions_awards" -dir ./model -c
goctl model mysql datasource -url="root:admin@tcp(127.0.0.1:3306)/datacenter" -table="app_questions_converts" -dir ./model -c
goctl model mysql datasource -url="root:admin@tcp(127.0.0.1:3306)/datacenter" -table="app_questions_lotteries" -dir ./model -c
goctl model mysql datasource -url="root:admin@tcp(127.0.0.1:3306)/datacenter" -table="app_questions_tests" -dir ./model -c

```