# Usage Guide (Admin)

## Log in
Open `/admin/login` and enter the admin password.

## Create an album
1. Go to **Albums**.
2. Click **New Album**.
3. Enter a title and description, then save.

## Upload photos
1. Open an album.
2. Use the upload form (single or batch).
3. After submitting, the UI will display a progress view while files are processed in the background.

Notes:
- Uploads are enqueued and processed asynchronously by a background worker. The upload handler saves incoming files to a temporary directory and returns a progress UI immediately.
- The progress UI polls the server and will update when processing completes. Wait for the progress indicator to reach 100% or click the provided "Refresh Album" button when the UI shows completion to see newly added photos.
- Temporary upload files are removed by the background worker after processing (or after a failed validation).

## Create a share link
1. Open the album or photo.
2. Click **Share**.
3. Set optional view limit and/or expiration time.
4. Copy the generated link.

## Manage share links
- Revoke a link to expire it immediately.
- View counts are tracked per unique viewer.

## Set album cover
Open an album and choose **Set Cover** on a photo.
