import { Router, Request, Response } from 'express';
import monitorRouter from './monitor';


function healthCheck(_req: Request, res: Response) {
    return res.send("Ok")
}

const baseRouter = Router();

/**
 * @openapi
 * /api/:
 *   get:
 *     description: API health check
 *     responses:
 *       200:
 *         description: Returns "Ok" string
 */
baseRouter.get("/", healthCheck);

baseRouter.use("/monitor", monitorRouter);

export default baseRouter;
