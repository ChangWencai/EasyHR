#!/usr/bin/env node
import { C as extensions, S as Container, b as accessibilityTarget, g as DOMPipe, h as EventSystem, m as FederatedContainer, x as AccessibilitySystem } from "./index.cjs";
import "./init-AxI8d1XI.cjs";

//#region ../../lib/pencil-editor/node_modules/pixi.js/lib/accessibility/init.mjs
extensions.add(AccessibilitySystem);
extensions.mixin(Container, accessibilityTarget);

//#endregion
//#region ../../lib/pencil-editor/node_modules/pixi.js/lib/events/init.mjs
extensions.add(EventSystem);
extensions.mixin(Container, FederatedContainer);

//#endregion
//#region ../../lib/pencil-editor/node_modules/pixi.js/lib/dom/init.mjs
extensions.add(DOMPipe);

//#endregion
export {  };