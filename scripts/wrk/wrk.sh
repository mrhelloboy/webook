# 登录

# 参数说明：
# -t1：表示使用1个线程
# -c2：表示使用2个连接
# -d1s：表示使用1秒钟的测试时间
# -s wrk_login.lua：表示使用wrk_login.lua脚本
# --latency：表示显示延迟信息
wrk -t1 -c2 -d1s -s login.lua --latency http://localhost:8080/user/loginJWT


# 注册
wrk -t1 -c2 -d1s -s signup.lua --latency http://localhost:8080/user/signup


# 获取用户信息
wrk -t2 -c2 -d1s -s profile.lua --latency http://localhost:8080/user/profileJWT