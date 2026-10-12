// Fixed-target local control ingress. ARTEX itself stays on the internal-only
// network. This is not a general HTTP proxy and never accepts target URLs.
import { createServer, request as forward } from 'node:http';

for (const [listenPort, targetPort] of [[8080, 8787], [8081, 18999]]) {
    createServer((request, response) => {
        const upstream = forward({ hostname: 'artex', port: targetPort, path: request.url,
            method: request.method, headers: request.headers, timeout: 40_000 }, (incoming) => {
            response.writeHead(incoming.statusCode, incoming.headers);
            incoming.pipe(response);
        });
        upstream.on('error', () => { if (!response.headersSent) response.writeHead(502); response.end(); });
        upstream.on('timeout', () => upstream.destroy());
        response.on('close', () => upstream.destroy());
        request.pipe(upstream);
    }).listen(listenPort, '0.0.0.0');
}
