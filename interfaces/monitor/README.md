# monitor 监视器

## 发行版一致性检查

在用户视角，检查线上发行版与 版本控制服务(VCS) 的最新交付版本是否一致


## 业务概念

Project - 项目
Component - 组件
ReleaseChannel - 发行渠道
VersionControlService - 版本控制服务
pagerduty - 一种事件管理平台
slack - 一种办公软件平台

### 运行

> ⚠️注意
> 
> 为了避免在命令行历史记录中遗留敏感信息，对于关键命令请使用空格 ` ` 开头，使 shell history 不记录当前行为


1. 准备 apple account 和 ipatool
   
   1. 注册一个 apple account，这个账号将被用于访问 appstore
   
   2. 下载 ipatool
   
   3. 在命令行使用 ` ipatool auth login -e <email> -p <password>` 使 ipa 可使用该账号访问 appstore5. 
   
   4. 执行 `ipatool purchase -b <bundle-id>` 订阅相关 app

   5. 测试。使用命令 `ipatool download -b <bundle id> -o .` 尝试下载相应app
      
      这时候会提示你创建一个 `keychain-passphrase`，请记好你设置的 `passphrase`，后续会用到

      在 mac 下，执行 `ipatool download`时，操作系统可能会问你是否始终信任该程序，如果你点击是，则后续运行一致性监控时，你不需要再输入该密码
   
2. 准备 google 账号
   
   1. 注册一个 google 账号
      
   2. 访问 [accounts.google.com/embedded/setup/v2/android](accounts.google.com/embedded/setup/v2/android)，
      执行登录过程直到最后页面卡住不动，从当前域名的 cookie 中获取 oauth_token
      
      注意！该 token 有效期 10min，且只能使用一次
   
   3. 执行子命令 ` main googleplay auth --code <token> --output ./token.txt` 以获取一个 access token 并记录至文件
      
      这个token file 就是后续访问 google play 的凭据

      [点此跳转至 googleplay 子命令的实现](../googleplay/googleplay.go)
   
   4. 测试。执行 `main googleplay download -b com.coinex.trade.play -t <前面生成的token file 的位置> -d <device.bin>`
   
3. 申请 github-api-token
   
   你可以申请 classic token，也可以申请细粒度 token，只要确保该 token 有相关 repo 的访问权限

4. 申请 slack-webhoook，并获取所需通知接收人的 slack user id
   
   点击 slack 用户名片右侧的三个点，然后点击“复制用户ID”即可
   
   [img_get_user_id_in_slack](../../resources/images/slack_query_user_id.png)

5. 启动服务
   
   ```bash
   main monitor release-consistency --github-api-token <github api token> --slack-id <slack id 1> --slack-id <slack id 2> --slack-webhook <slack webhook> --gplay-token <googleplay token filename> --gplay-device device.bin
   ```
   
   更多参数可执行 `main monitor releaseconsistency -h` 获取


### 主要检查方式：

- 检查关键代码

    如交易所前端的 `_APP_VERSION` 字段，或 IPA 中的 UUID 字段

- 检查关键文件的校验码
    
    如对于 `.apk`，解压缩后，检查 `*.dex`、`*.so`、`*.arsc` 的校验码

## 其他说明

1. .apk 和 .ipa 文件本质都是 zip，zip文件自带 CRC 码。这意味着：
   
   1. 可以用 unzip 解压
   
   2. 可以直接使用 CRC 比较特定文件的一致性
   
2. [googleplay核心代码](../../infrastructure/googleplay), 参考[github:3052/google](https://github.com/3052/google) 实现 —— 主要对无用代码进行了删减

3. google play 的 access token 不是 open api 的 token，访问接口也是 app 端使用的 rpc 协议
   
   这意味着目前的访问方式并不可靠，未来如果协议发生变更，可能会导致 google play 渠道的检查程序无法正常运行
   
   可通过以下 repo 的 issue 关注相关问题
   
   - https://github.com/3052/google
   
   - https://github.com/NoMore201/googleplay-api
   
   - https://github.com/egirault/googleplay-api

4. 关于 ipatool 如果遇到困难，可尝试在 https://github.com/majd/ipatool/issues 获取帮助，并将问题在此记录便于其他成员了解情况
   