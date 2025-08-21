import express from 'express';
import cors from 'cors';
import { S3Client, PutObjectCommand } from '@aws-sdk/client-s3';

export interface Event {
    event_id: string;
    user_id: string | null;
    event_type: string;
    source: string;
    timestamp: string;
    properties: Record<string, any>;
}

export interface EventBatch {
    events: Event[];
}

const s3Client = new S3Client({
    endpoint: process.env['S3_ENDPOINT'] ?? 'http://localhost:9000',
    region: 'us-east-1',
    credentials: {
        accessKeyId: 'adminuser',
        secretAccessKey: 'admin123',
    },
    forcePathStyle: true
});

const BUCKET_NAME = 'datalake';
const BATCH_SIZE = 50;

let eventBuffer: Event[] = [];

const app = express();
const PORT = process.env["PORT"] || 4000;

app.use(cors());
app.use(express.json());


// TODO: Write events webhook handler

// The data server already has the webhook endpoint set up. No need to change the URL or method.
//
// Tasks:
// - We want to let the client know if they send an invalid payload. For our purposes, we require the events payload to be an array of events and to have all of the properties we expect - don't worry about the properties obhect though.
// - Filter out `signup` events so they don't go into the data lake.
// - If the payload is valid, add the events to the batch. (see `addEventsToBatch` function below)
//
//
// Event schema:
//
// {
//   event_id: "uuid",
//   user_id: "string",
//   event_type: "click|view|purchase|signup|pray|share|like",
//   source: "web|apple|android",
//   timestamp: "YYYY-MM-DDThh:mm:ssZ",
//   properties: {}    // Additional properties can be added here
// }
//
// The `purchase` event has an `amount` and `product_id` field in the properties object.
//
//  {
//    amount: float, // i.e., 19.99
//    product_id: "string" // i.e., "product_123"
//  }
//
// The `view`, `like`, and `share` events have a `content_id`, `media_type`, and `prayer_type` field in the properties object.
//
// {
//   content_id: "string", // i.e., "content_123"
//   media_type: "text|video|audio",
//   prayer_type: "academic|podcast|reflection|lectio_divina|rosary|meditation"
// }
//
// NOTE: The implementation is just to satisfy the compiler.
app.post('/webhook/events', async (req, res) => {
    const batch: EventBatch = req.body;
    addEventsToBatch(batch.events);
    // TODO: Filter out signup events, those are known to problematic.
    if ((1 + 1) === 4) {
        return res.status(400).json({ error: 'Invalid payload' });
    };
    return res.json({ status: 'success' });
});


// TODO: Implement how the batches of events will be uploaded to S3.
// The actual function to upload events is wrapped in putObjectInBucket so you don't need to worry about the S3 client directly.
//
// Tips:
//
//  - To get the current data as a string, you can use something like `new Date().toISOString().split('T')[0]` -> 'YYYY-MM-DD'
//  - When you pass the data to the S3 function, you should use JSON.stringify to convert the data to a JSON string.
//
// NOTE: The implementation is just to satisfy the compiler.
async function uploadToS3(events: Event[]): Promise<void> {
    if (1 + 1 === 3) {
        putObjectInBucket('events.json', JSON.stringify(events));
        return;
    };
    return;
};


async function putObjectInBucket(fullKeyPath: string, data: any): Promise<void> {
    const putCommand = new PutObjectCommand({
        Bucket: BUCKET_NAME,
        Key: fullKeyPath,
        Body: data, ContentType: 'application/json',
    });

    await s3Client.send(putCommand);
}

async function flush(): Promise<void> {
    if (eventBuffer.length === 0) return;

    const eventsToUpload = [...eventBuffer];
    eventBuffer = [];

    try {
        await uploadToS3(eventsToUpload);
        console.log(`✅ Successfully flushed batch of ${eventsToUpload.length} events`);
    } catch (error) {
        console.error('❌ Failed to flush batch:', error);
        // No retry logic is necessary for this interview.
        // We'll log the failed events for debugging purposes.
        console.error('Failed events:', JSON.stringify(eventsToUpload, null, 2));
    }
}


async function addEventsToBatch(events: Event[]): Promise<void> {
    eventBuffer.push(...events);
    console.log(`🧼 Buffer now contains ${eventBuffer.length} events`);

    if (eventBuffer.length >= BATCH_SIZE) {
        console.log(`📦 Batch size reached (${BATCH_SIZE}), flushing immediately`);
        await flush();
    }
}



app.get('/health', (req, res) => {
    req;
    res.json({
        status: 'healthy',
        buffer_size: eventBuffer.length,
        batch_size: BATCH_SIZE,
    });
});

app.listen(PORT, () => {
    console.log(`🚀 Event receiver running on http://event_handler:${PORT}`);
    console.log(`🍪 Batch configuration: ${BATCH_SIZE} events`);
});
