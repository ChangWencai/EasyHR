#!/usr/bin/env node
import { C as extensions, _ as GraphicsContextSystem, a as BitmapTextPipe, c as GpuParticleContainerPipe, d as GraphicsPipe, f as FilterSystem, i as HTMLTextPipe, l as GlParticleContainerPipe, n as CanvasTextPipe, o as TilingSpritePipe, p as FilterPipe, r as HTMLTextSystem, s as NineSliceSpritePipe, t as CanvasTextSystem, u as MeshPipe, v as TickerPlugin, y as ResizePlugin } from "./index.cjs";

//#region ../../lib/pencil-editor/node_modules/pixi.js/lib/app/init.mjs
extensions.add(ResizePlugin);
extensions.add(TickerPlugin);

//#endregion
//#region ../../lib/pencil-editor/node_modules/pixi.js/lib/scene/graphics/init.mjs
extensions.add(GraphicsPipe);
extensions.add(GraphicsContextSystem);

//#endregion
//#region ../../lib/pencil-editor/node_modules/pixi.js/lib/scene/mesh/init.mjs
extensions.add(MeshPipe);

//#endregion
//#region ../../lib/pencil-editor/node_modules/pixi.js/lib/scene/particle-container/init.mjs
extensions.add(GlParticleContainerPipe);
extensions.add(GpuParticleContainerPipe);

//#endregion
//#region ../../lib/pencil-editor/node_modules/pixi.js/lib/scene/text/init.mjs
extensions.add(CanvasTextSystem);
extensions.add(CanvasTextPipe);

//#endregion
//#region ../../lib/pencil-editor/node_modules/pixi.js/lib/scene/text-bitmap/init.mjs
extensions.add(BitmapTextPipe);

//#endregion
//#region ../../lib/pencil-editor/node_modules/pixi.js/lib/scene/text-html/init.mjs
extensions.add(HTMLTextSystem);
extensions.add(HTMLTextPipe);

//#endregion
//#region ../../lib/pencil-editor/node_modules/pixi.js/lib/scene/sprite-tiling/init.mjs
extensions.add(TilingSpritePipe);

//#endregion
//#region ../../lib/pencil-editor/node_modules/pixi.js/lib/scene/sprite-nine-slice/init.mjs
extensions.add(NineSliceSpritePipe);

//#endregion
//#region ../../lib/pencil-editor/node_modules/pixi.js/lib/filters/init.mjs
extensions.add(FilterSystem);
extensions.add(FilterPipe);

//#endregion
export {  };