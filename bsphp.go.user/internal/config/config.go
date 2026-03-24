// Package config 存放 BSPHP 账号模式演示程序必须改动的服务端地址、密钥与 Web 链接。
//
// 【本程序行为说明】
// 1) 界面为跨平台图形库绘制，与系统原生控件外观可能略有不同。
// 2) Web 登录：在系统浏览器打开登录页，后台每 2 秒请求心跳；返回码 5031 时视为登录成功并打开控制台。
// 3) 机器码：优先读取本机 HostID；失败则生成随机 ID 并写入用户配置目录后长期复用。
// 4) 验证码图：HTTP 加载，刷新时在地址后追加 &_=时间戳 避免缓存。
package config

// --- 必须配置（与后台「当前应用」AppEn 地址、通信密钥、RSA 保持一致）---

// BSPHPURL AppEn 完整入口 POST 地址
const BSPHPURL = "https://demo.bsphp.com/AppEn.php?appid=8888888&m=95e87faf2f6e41babddaef60273489e1&lang=0"

// BSPHPMutualKey 通信 mutualkey
const BSPHPMutualKey = "6600cfcd5ac01b9bb3f2460eb416daa8"

// BSPHPServerPrivateKey 服务器私钥 Base64（解密响应、派生请求 AES）
const BSPHPServerPrivateKey = "MIIEqAIBADANBgkqhkiG9w0BAQEFAASCBJIwggSOAgEAAoH+DEr7H5BhMwRA9ZWXVftcCWHznBdl0gQBu5617qSe9in+uloF1sC64Ybdc8Q0JwQkGQANC5PnMqPLgXXIfnl7/LpnQ/BvghQI5cr/4DEezRKrmQaXgYfXHL3woVw7JIsLpPTGa7Ar9S6SEH8RcPIbZjlPVRZPwV3RgWgox2/4lkXsmopqD+mEtOI/ntvti147nEpK2c7cdtCU5M2hQSlIXsTWvri88RTYJ/CtopBOXarUkNBfpWGImiYGsmbZI+YZ6uU0wSYlq8huu+pkTseUUiymzmv8Rpg3coi7YU+pszvB9wnQ1Rz6Z/B6Z3WN7d6OP7f9w0Q0WvgrsKcEJhMCAwEAAQKB/gHa5t6yiRiL0cm902K0VgVMdNjfZww0cpZ/svDaguqfF8PDhhIMb6dNFOo9d6lTpKbpLQ7MOR2ZPkLBJYqAhsdy0dac2BcHMviKk+afQwirgp3LMt3nQ/0gZMnVA0/Wc+Fm1vK1WUzcxEodAuLKhnv8tg4fGdYSdGVU9KJ0MU1bKQZXv0CAIhJYWsiCa5y5bFO7K+ia+UIVBHcvITQLzlgEm+Z/X6ye5cws4pWbk8+spsBDvweb5jpelbkCYs5C5TRNIWXk7+QxTXTg1vrcsmZRcmpRJq7sOd3faZltNHTIlB3HhWnsf47Bz334j9RtU8iqonbuBmcnYbD3+bvBAn891RGdAl+rVU/sJ2kPXmV4eqJOwJfbi8o1WYDp4GcK0ThjrZ1pmaZMj2WTjb3QX1VUoi+7l3389KzzDn0VXLKXZvGxmLikA1FWuuLUmwfNTxyxtGTBVeZCEaQ2lEJuaDGsK0oLi4Bo8ELfQw6JFK7jlgtTlflcYcul99P9BThDAn8y5TpSQy8/07LCgMMZOgJomYzQUmd14Zn2VQLH1u1Z4v2CPlOzGanDt7mmGZCew7iMSO1P0TrwDIreKzYyERuVvZti/IFHH1+J1hAbvk9SJGmdt46W5lyIp3xjdR2QmiK+hSsc8HF9R+zPaSe9yGA8+FwxLRfo0snGP3MC3aXxAn4n2iyABgejZlkc3EnanfzIqkHygC9gUbkCqa1tEDVZw3+Uv1G1vlJxBftyHuk4ZDmbUu1w+zM41nqiLbRxEE4LR06AKO7Yx0qlm86XOVTN/y9/WcWW1saRzs0IYIZwordhQIV463DYMgLn41B7Cdmu1gZ22TLfWCjpz9HSQosCfwMJu9l9OSzOLjV+CidPVyV3RPiKcrKOrOoPWQMkyTY8XnWP0t82APQ121cW35Mai8GT+NZy3tnFZeStH6cNbmAZ2VSnTfA45zMLHBsL2SBGHCfV9ST8yzk9BifJreIb0UceG9y2XY/k4zXeSQkDFPuOt7IXxv2W14SF9Q+Ou4ECfzfRP1hXPwq2w4YJ8sLmqWJT+3aMDucei5MJEAJNifZWhdW0GIrlKRSbhIgLAunxq+KK+mAPqqWw7Prsa21JbXSe3gugusu5d6ESURvLENRKI+Pp9TgRESsydeLy8VcPKRJ5/Ct7/p6QB3A+7F/iPNE2GagGffG9i7e+OdcToYQ="

// BSPHPClientPublicKey 客户端公钥 Base64（加密请求 RSA 段）
const BSPHPClientPublicKey = "MIIBHjANBgkqhkiG9w0BAQEFAAOCAQsAMIIBBgKB/g26m2hYtESqcKW+95Lr+PfCd4bwHW2Z+mM0/vcKQ5j/ZGMigqkgl3QXCEcsCaw0KFSmqAPtLbrl6p5Sp+ZUSYEYQhSxAajE5qRCd3k0r/MIQQanBaOALkP71/u6U2SZhrTXd05n1wQo6ojMH/xVunBOFOa/Eon/Y5FVh6GiJpwwDkFzTlnecmff7Y+VDqRhZ7vu2CQjApOx23N6DiFEmVZYEb/efyASngoZ+3A/DSB5cwbaYVZ21EhPe/GNcwtUleFHn+d4vb0cvolO3Gyw6ObceOT/Q7E3k8ejIml6vPKzmRdtw0FXGOJTclx1CjShRDfXoUjFGyXHy3sZs9VLAgMBAAE="

// BSPHPCodeURL 图片验证码前缀（后接 BSphpSeSsL）
const BSPHPCodeURL = "https://demo.bsphp.com/index.php?m=coode&sessl="

// --- WebAPI 用系统浏览器打开（Web 登录、控制台续费购买等）---

const BSPHPWebLoginURL = "https://demo.bsphp.com/index.php?m=webapi&c=software_auth&a=index&daihao=8888888&BSphpSeSsL="

const BSPHPRenewURL = "https://demo.bsphp.com/index.php?m=webapi&c=salecard_renew&a=index&daihao=8888888"

const BSPHPRenewCardURL = "https://demo.bsphp.com/index.php?m=webapi&c=salecard_gencard&a=index&daihao=8888888"

const BSPHPRenewStockCardURL = "https://demo.bsphp.com/index.php?m=webapi&c=salecard_salecard&a=index&daihao=8888888"
