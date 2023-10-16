package draylix.handler;

import draylix.server.Visitor;
import io.netty.buffer.ByteBuf;
import io.netty.channel.ChannelHandler;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.InetSocketAddress;

@ChannelHandler.Sharable
public class HttpsProxyInitHandler extends SimpleChannelInboundHandler<ByteBuf> {
    private Logger logger= LoggerFactory.getLogger(this.getClass());
    public static final HttpsProxyInitHandler INSTANCE=new HttpsProxyInitHandler();
    @Override
    protected void channelRead0(ChannelHandlerContext ctx, ByteBuf buf) throws Exception {
        byte[] reqb=new byte[buf.readableBytes()];
        buf.readBytes(reqb);
        String req = new String(reqb);

        InetSocketAddress addr=parseConnectReq(req);
        if (addr==null) {
            ctx.writeAndFlush(wrap(ctx,"HTTP/1.0 502 Bad Gateway\r\n\r\n"));
            ctx.close();
            return;
        }

        Visitor visitor = new Visitor(ctx.channel());
        if (!visitor.connect(addr)) {
            ctx.writeAndFlush(wrap(ctx,"HTTP/1.0 502 Bad Gateway\r\n\r\n"));
            ctx.close();
            return;
        }

        ctx.writeAndFlush(wrap(ctx,"HTTP/1.0 200 OK\r\n\r\n")).addListener(f->{
            logger.info(" {} -> httpsProxy -> {}",ctx.channel().remoteAddress().toString(),addr.getAddress().toString());
            ctx.pipeline().addLast(new TransmitHandler(visitor));
            ctx.pipeline().remove(this);
        });
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) throws Exception {
        logger.error(cause.getMessage());
    }

    private ByteBuf wrap(ChannelHandlerContext ctx ,String s){
        byte[] bytes = s.getBytes();
        ByteBuf buffer = ctx.alloc().buffer(bytes.length);
        buffer.writeBytes(bytes);
        return buffer;
    }

    private InetSocketAddress parseConnectReq(String req) {
        String[] lines=req.split("\r\n");
        String[] args=lines[0].split(" ");
        if (args.length<3) return null;
        if (!"CONNECT".equals(args[0])) return null;
        String[] hostPort = args[1].split(":");
        if(hostPort.length<2) return null;
        String host;
        int port;
        host=hostPort[0];
        try {
            port=Integer.parseInt(hostPort[1]);
        }catch (NumberFormatException e){
            return null;
        }
        return new InetSocketAddress(host,port);
    }
}
