# manageChain
一、目的与初衷
1、基于fabric1.2版本对外提供restful api服务，能够实现对链的管理操作，旨在提供便捷简单创建联盟链的服务，管理联盟链服务，推进联盟链生态建设；
2、主要提供生成创建链的秘钥证书文件，创世块接口；创建链；chaincode操作相关接口；
3、提供升级链内成员组织接口，主要添加链内成员、删除链内成员接口；
4、支持国密和原生密码学服务；

二、应用
1、首先可以通过bee run运行该程序，通过channel/channel_test.go里面的测试用例生成秘钥证书文件、创世块，然后可以采用docker-compose启动区块链节点；
2、通过channel_test.go测试用例，进行链的创建，合约安装，实例化合约等；
3、当联盟成员发生变化，例如需要增加成员，删除成员可以通过channel/channel_test.go用例来进行对配置块进行升级;这里需要注意链的adminpolicy,可以majority\any\all.

三、后续计划
1、支持最新版本fabric,一些SDK接口发生变化
