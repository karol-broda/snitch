# Demo Recording

This directory contains files for recording the snitch demo GIF in a controlled Docker environment.

## Files

- `Dockerfile` - builds snitch and sets up fake network services
- `demo.tape` - VHS script that records the demo
- `entrypoint.sh` - starts fake services before recording

## Recording the Demo

From the project root:

```bash
# build the demo image
docker build -f demo/Dockerfile -t snitch-demo .

# run and output demo.gif to this directory
docker run --rm -v $(pwd)/demo:/output snitch-demo
```

The resulting `demo.gif` will be saved to this directory.

## Fake Services

The container runs several fake services to demonstrate snitch:

| Service | Port | Protocol |
|---------|------|----------|
| nginx   | 80   | TCP      |
| web app | 8080 | TCP      |
| node    | 3000 | TCP      |
| postgres| 5432 | TCP      |
| redis   | 6379 | TCP      |
| mongo   | 27017| TCP      |
| mdns    | 5353 | UDP      |
| ssdp    | 1900 | UDP      |

Plus some simulated established connections between services.

## Customizing

Edit `demo.tape` to change what's shown in the demo. See [VHS documentation](https://github.com/charmbracelet/vhs) for available commands.

