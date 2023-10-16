package draylix.handler;

import draylix.server.Visitor;
import io.netty.buffer.ByteBuf;
import io.netty.channel.ChannelHandler;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.InetSocketAddress;
import java.net.MalformedURLException;
import java.net.URL;
import java.util.StringJoiner;

@ChannelHandler.Sharable
public class HttpProxyInitHandler extends SimpleChannelInboundHandler<String> {
    public static final HttpProxyInitHandler INSTANCE=new HttpProxyInitHandler();
    private Logger logger= LoggerFactory.getLogger(this.getClass());

    @Override
    protected void channelRead0(ChannelHandlerContext ctx, String httpReq) throws Exception {
        InetSocketAddress addr = getAddr(httpReq);
        if (addr==null){
            ctx.writeAndFlush(wrap(ctx,"HTTP/1.0 502 Bad Gateway\r\n\r\n"));
            ctx.close();
            return;
        }
        Visitor visitor=new Visitor(ctx.channel());
        if (!visitor.connect(addr)) {
            ctx.writeAndFlush(wrap(ctx,"HTTP/1.0 502 Bad Gateway\r\n\r\n"));
            ctx.close();
            return;
        }
        ctx.pipeline()
                .addLast(new TransmitHandler(visitor))
                .remove("StringDecoder");
        visitor.send(wrap(ctx,changeHttpReq(httpReq)));
        logger.info(" {} -> httpProxy -> {}",ctx.channel().remoteAddress().toString(),addr.getAddress().toString());
        ctx.pipeline().remove(this);
    }

    private String changeHttpReq(String httpReq) throws MalformedURLException {
        String[] args = httpReq.split(" ");
        URL url=new URL(args[1]);
        args[1]=url.getPath();
        StringJoiner sj=new StringJoiner(" ");
        for (int i = 0; i < args.length; i++) {
            sj.add(args[i]);
        }
        return sj.toString();
    }

    public InetSocketAddress getAddr(String httpReq) {
        String[] args = httpReq.split(" ");
        if (args.length<2){
            return null;
        }
        URL url;
        try {
            url=new URL(args[1]);
        } catch (MalformedURLException e) {
            return null;
        }
        int port=url.getPort();
        if (port==-1){
            port = args[1].startsWith("http") ? 80 : 443;
        }

        return new InetSocketAddress(url.getHost(),port);
    }

    private ByteBuf wrap(ChannelHandlerContext ctx ,String s){
        byte[] bytes = s.getBytes();
        ByteBuf buffer = ctx.alloc().buffer(bytes.length);
        buffer.writeBytes(bytes);
        return buffer;
    }
}
