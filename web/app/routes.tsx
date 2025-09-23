import { createBrowserRouter, type RouteObject, type ShouldRevalidateFunction, type ShouldRevalidateFunctionArgs } from "react-router";

import { flatRoutes } from "@react-router/fs-routes";
import type { RouteConfig } from "@react-router/dev/routes";

export default flatRoutes() satisfies RouteConfig;

