const { post } = require('./request')

function sendVerifyCode(phone) {
  return post('/wxmp/auth/send-code', { phone })
}

function countDown(setSeconds, setter) {
  let left = setSeconds
  setter(left)
  const timer = setInterval(() => {
    left--
    if (left <= 0) { clearInterval(timer); setter(0) }
    else setter(left)
  }, 1000)
  return () => clearInterval(timer)
}

module.exports = { sendVerifyCode, countDown }
