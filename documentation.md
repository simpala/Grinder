# Grinder Project Documentation

## 1. Architecture Review

The "Grinder" renderer is a voxel-based software renderer built on a unique volume-dicing architecture. Unlike traditional rasterizers or ray tracers, Grinder treats the view frustum as a solid volume to be recursively subdivided. This approach is centered around the **Axis-Aligned Bounding Box (AABB)**, which serves as the universal primitive for all geometry.

### Core Concepts

*   **Universal Primitive:** All geometric shapes implement a `Contains(Point)` and `Intersects(AABB)` interface, unifying the rendering algorithm.
*   **Lazy Dicing:** Geometry is only subdivided when it is visible within a screen-space volume, providing a natural Level of Detail (LOD).
*   **Painter's Algorithm:** The renderer sorts shapes from furthest to closest to handle occlusion, eliminating the need for a traditional Z-buffer.
*   **Tiled Rendering:** The screen is divided into tiles that can be rendered in parallel, making the architecture "embarrassingly parallel."

### Architectural Issues

The current implementation shows signs of architectural drift. The most significant issue is the disabled AABB culling in `renderer.go`, which indicates that the intersection tests are not reliable for all transformations (e.g., rotations). The comment `Do not cull. AABB culling is incorrect for rotated shapes` suggests a critical flaw in the core algorithm.

## 2. Rendering Pipeline

The rendering process is divided into two main passes:

1.  **Pass 1: Dicing (Visibility)**
    *   The renderer starts with a single AABB representing the entire view frustum.
    *   This AABB is recursively subdivided into eight smaller boxes.
    *   At a minimum size threshold, a fine-grind search is performed to find the closest surface for each pixel.
    *   A `SurfaceData` buffer is populated with the hit point, normal, shape, and depth for each pixel.

2.  **Pass 2: Shading**
    *   The renderer iterates over the `SurfaceData` buffer.
    *   For each hit point, it performs shading calculations with 9x Super-Sampled Anti-Aliasing (SSAA).
    *   For soft shadows, the light position is jittered for each sample.

## 3. Organizational Improvement Suggestions

The project is well-structured, but the following improvements would enhance maintainability:

*   **Split `shape.go`:** The `pkg/geometry/shape.go` file contains multiple shape implementations. Splitting each shape into its own file (e.g., `sphere.go`, `plane.go`) would improve organization.
*   **Create an `internal` Package:** Code that is not intended for public use should be moved to an `internal` package to enforce encapsulation.
*   **Add a Test Suite:** The project currently lacks a test suite. Adding unit tests for the `math` and `geometry` packages and integration tests for the renderer would improve stability.

## 4. Proposed Roadmap for New Tasks

The following is a proposed roadmap for future development:

### Task 1: Refactor the Renderer

*   **Clean up `renderer.go`:** Remove the commented-out `Render` function and other dead code.
*   **Add Comments:** Add extensive comments to the `subdivide` function to clarify the rendering algorithm.
*   **Improve Readability:** Refactor the shading pass to improve readability and maintainability.

### Task 2: Fix AABB Culling

*   **Implement Correct Intersection Tests:** Implement correct AABB intersection tests for all shape types, including rotated shapes.
*   **Re-enable Culling:** Re-enable AABB culling in the `subdivide` function to improve performance.

### Task 3: Implement a Test Suite

*   **Unit Tests:** Add unit tests for the `math` and `geometry` packages to ensure correctness.
*   **Integration Tests:** Add integration tests that render a scene and compare the output to a reference image.

### Task 4: Add New Features

*   **Rotated Shapes:** Add support for rotated shapes by implementing proper transformations.
*   **Triangle Mesh Loader:** Add a loader for triangle meshes (e.g., in OBJ or STL format).
*   **Advanced Materials:** Add support for advanced materials with properties like roughness, metalness, and emission.
