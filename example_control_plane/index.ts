import express, {Request} from 'express'

const app = express()
app.use(express.json())

interface Key {
    UserID: string
    KeyID: string
    SecretKey: string
}

app.get('/key/:keyID', async (req: Request<{keyID: string}, Key>, res) => {
    console.log("serving key", req.params.keyID)
    res.json({
        KeyID: "testuser",
        SecretKey: "testpassword",
        UserID: "testid"
    })
})

app.post('/', async (req: Request, res) => {
    console.log("got req for root")
    res.json({
        KeyID: "testuser",
        SecretKey: "testpassword",
        UserID: "testid"
    })
})

app.listen('8888', () => {
    console.log('listening on port 8888')
})
