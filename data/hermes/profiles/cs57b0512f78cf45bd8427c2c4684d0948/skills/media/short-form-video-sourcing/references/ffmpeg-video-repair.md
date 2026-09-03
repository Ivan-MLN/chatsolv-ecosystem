# FFmpeg Video Repair & Playback Fix

When downloading web/CDN MP4 videos (e.g., slicedrive, Twitter-vork muxer, etc.), video files can sometimes have corrupted NAL unit headers or non-standard H.264 stream muxing. This causes WhatsApp or mobile media players to play audio fine while the video stream freezes/stucks after a few seconds.

## 1. Diagnosis
Check for H.264 NAL unit errors or frame decoding issues:

```bash
ffmpeg -v error -i input.mp4 -f null -
```

**Common error signatures:**
- `[h264 @ ...] Invalid NAL unit size (1695 > 1483).`
- `[h264 @ ...] Error splitting the input into NAL units.`
- `Error while decoding stream #0:0: Invalid data found when processing input`

## 2. Re-encoding & Repair Command
Re-encode the video stream to standard H.264 with `yuv420p` pixel format and `+faststart` MOOV atom placement while preserving audio:

```bash
ffmpeg -y -i input.mp4 -c:v libx264 -preset fast -crf 22 -pix_fmt yuv420p -movflags +faststart -c:a copy output_fixed.mp4
```

## 3. Fast Compression & Buffer Optimization for Mobile/WhatsApp
If the video is large (e.g. >30MB) or buffers slowly and freezes midway on mobile:
- Reduce resolution (e.g. max 720p width), set keyframes per 2 seconds (GOP 60 for 30fps), increase CRF (26), and encode audio to AAC:

```bash
ffmpeg -y -i input.mp4 -vf "scale='min(720,iw)':-2" -c:v libx264 -preset medium -crf 26 -g 60 -keyint_min 60 -sc_threshold 0 -pix_fmt yuv420p -movflags +faststart -c:a aac -b:a 128k output_compressed.mp4
```

## 4. Verification
Verify that the output file decodes cleanly without error logs:

```bash
ffmpeg -v error -i output_fixed.mp4 -f null -
```
If stdout/stderr is clean, the video file is safe to send via `wa_send_file`.
