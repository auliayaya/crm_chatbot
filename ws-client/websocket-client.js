// websocket-client.js
const WebSocket = require('ws')
const readline = require('readline')

// Create readline interface for manual input
const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
})

// Configuration
const config = {
  url: 'ws://localhost:8080/chat/ws',
  userId: 'agent1',
  customerId: 'cust123',
  autoMessageInterval: 3000, // ms
  autoMessageCount: 5, // send 5 messages then stop
}

let ws
let interval
let messageCount = 0

// Connect to the WebSocket server
function connect() {
  const url = `${config.url}?user_id=${config.userId}&customer_id=${config.customerId}`
  console.log(`Connecting to ${url}...`)

  ws = new WebSocket(url)

  ws.on('open', () => {
    console.log('Connected!')
    showPrompt()
  })

  ws.on('message', (data) => {
    try {
      const message = JSON.parse(data)
      console.log(`\nReceived: ${JSON.stringify(message)}`)
    } catch (e) {
      console.log(`\nReceived: ${data}`)
    }
    showPrompt()
  })

  ws.on('close', () => {
    console.log('Disconnected')
    stopAutoMessages()
    process.exit(0)
  })

  ws.on('error', (error) => {
    console.error(`Error: ${error.message}`)
  })
}

// Send a message
function sendMessage(content) {
  if (!ws || ws.readyState !== WebSocket.OPEN) {
    console.error('Not connected')
    return
  }

  const message = {
    content: content,
    type: 'user_message',
  }

  ws.send(JSON.stringify(message))
  console.log(`Sent: ${JSON.stringify(message)}`)
}

// Start sending automated messages
function startAutoMessages() {
  if (interval) {
    console.log('Automation already running')
    return
  }

  messageCount = 0
  console.log(
    `Starting automated messages every ${config.autoMessageInterval}ms`
  )

  interval = setInterval(() => {
    messageCount++
    sendMessage(`Automated message #${messageCount}`)

    if (messageCount >= config.autoMessageCount) {
      stopAutoMessages()
    }
  }, config.autoMessageInterval)
}

// Stop automated messages
function stopAutoMessages() {
  if (interval) {
    clearInterval(interval)
    interval = null
    console.log('Stopped automated messages')
  }
}

// Show command prompt
function showPrompt() {
  rl.question('> ', (input) => {
    if (input === 'exit' || input === 'quit') {
      if (ws) ws.close()
      rl.close()
      process.exit(0)
    } else if (input === 'auto') {
      startAutoMessages()
      showPrompt()
    } else if (input === 'stop') {
      stopAutoMessages()
      showPrompt()
    } else if (input) {
      sendMessage(input)
    } else {
      showPrompt()
    }
  })
}

console.log('WebSocket Client')
console.log(
  "Commands: 'auto' to start auto-messages, 'stop' to stop, 'exit' to quit"
)
connect()
