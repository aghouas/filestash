// plugins/plg_stl_viewer/plugin.go
package plg_stl_viewer

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mickael-kerjean/filestash/server/common"
)

func init() {
    // Register the custom opener for .stl files
    common.Hooks.Register.XDGOpen(`
        if (mime === "model/stl") {
            return ["appframe", {"endpoint": "/stl-viewer"}];
        }
    `)

    // Register the HTTP endpoint to serve the iframe content
    common.Hooks.Register.HttpEndpoint(registerSTLViewerEndpoint)
}

func registerSTLViewerEndpoint(r *mux.Router, app *common.App) error {
    r.HandleFunc(
        "/stl-viewer",
        STLViewerHandler,
    ).Methods("GET")
    return nil
}

func STLViewerHandler(res http.ResponseWriter, req *http.Request) {
    // Extract the file path from the query parameters
    path := req.URL.Query().Get("path")
    if path == "" {
        http.Error(res, "File path is required", http.StatusBadRequest)
        return
    }

    // Serve the HTML content
    res.Header().Set("Content-Type", "text/html")
    res.WriteHeader(http.StatusOK)

    // Build the iframe HTML content with the STL viewer
    htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <title>STL Viewer</title>
    <style>
        body { margin: 0; }
        #viewer { width: 100%; height: 100vh; }
    </style>
    <script type="importmap">
    {
        "imports": {
            "three": "https://cdn.jsdelivr.net/npm/three@0.169.0/build/three.module.js",
            "three/addons/": "https://cdn.jsdelivr.net/npm/three@0.169.0/examples/jsm/"
        }
    }
    </script>
</head>
<body>
    <div id="viewer"></div>
    <script type="module">
        import * as THREE from 'three';
        import { OrbitControls } from 'three/addons/controls/OrbitControls.js';
        import { STLLoader } from 'three/addons/loaders/STLLoader.js';

        // JavaScript code to load and render the .stl file
        const filePath = "` + path + `";
        const fileUrl = "/api/files/cat?path=" + encodeURIComponent(filePath);

        // Set up the scene, camera, and renderer
        const scene = new THREE.Scene();
        const camera = new THREE.PerspectiveCamera(70, window.innerWidth / window.innerHeight, 0.1, 1000);
        const renderer = new THREE.WebGLRenderer({ antialias: true });
        renderer.setSize(window.innerWidth, window.innerHeight);
        document.getElementById('viewer').appendChild(renderer.domElement);

        // Add controls
        const controls = new OrbitControls(camera, renderer.domElement);

        // Add lighting
        const ambientLight = new THREE.AmbientLight(0x404040, 2);
        scene.add(ambientLight);

        const directionalLight = new THREE.DirectionalLight(0xffffff, 1);
        directionalLight.position.set(1, 1, 1).normalize();
        scene.add(directionalLight);

        // Load the STL file
        const loader = new STLLoader();
        fetch(fileUrl, { credentials: 'include' })
            .then(response => {
                if (!response.ok) {
                    throw new Error('Network response was not ok. Status: ' + response.status + ' ' + response.statusText);
                }
                return response.arrayBuffer();
            })
            .then(buffer => {
                const geometry = loader.parse(buffer);
                const material = new THREE.MeshPhongMaterial({ color: 0x5588aa, specular: 0x111111, shininess: 200 });
                const mesh = new THREE.Mesh(geometry, material);
                scene.add(mesh);

                // Center the mesh
                const middle = new THREE.Vector3();
                geometry.computeBoundingBox();
                geometry.boundingBox.getCenter(middle);
                mesh.position.set(-middle.x, -middle.y, -middle.z);

                // Adjust camera position
                const size = new THREE.Vector3();
                geometry.boundingBox.getSize(size);
                const maxDim = Math.max(size.x, size.y, size.z);
                const fov = camera.fov * (Math.PI / 180);
                let cameraZ = Math.abs(maxDim / 2 / Math.tan(fov / 2));
                camera.position.z = cameraZ * 1.5;

                // Update controls and render
                controls.target.set(0, 0, 0);
                controls.update();
                animate();
            })
            .catch(error => {
                console.error("Error loading STL file:", error);
                alert("Error loading STL file: " + error.message);
            });

        // Handle window resize
        window.addEventListener('resize', function() {
            const width = window.innerWidth;
            const height = window.innerHeight;
            renderer.setSize(width, height);
            camera.aspect = width / height;
            camera.updateProjectionMatrix();
        });

        // Animation loop
        function animate() {
            requestAnimationFrame(animate);
            controls.update();
            renderer.render(scene, camera);
        }
    </script>
</body>
</html>
    `

    res.Write([]byte(htmlContent))
}
