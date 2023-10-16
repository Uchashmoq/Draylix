package draylix.handler;

import draylix.server.Visitor;
import io.netty.channel.ChannelHandler;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.codec.socksx.SocksVersion;
import io.netty.handler.codec.socksx.v5.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

@ChannelHandler.Sharable
public class Socks5CommandRequestHandler extends SimpleChannelInboundHandler<DefaultSocks5CommandRequest> {
    private final Logger logger = LoggerFactory.getLogger(this.getClass());
    public static final Socks5CommandRequestHandler INSTANCE=new Socks5CommandRequestHandler();
    public Socks5CommandRequestHandler(){}

    @Override
    protected void channelRead0(ChannelHandlerContext ctx, DefaultSocks5CommandRequest msg)  {
        if(msg.decoderResult().isFailure()){
            logger.error(msg.decoderResult().cause().getMessage());
            ctx.fireChannelRead(msg);
            return;
        }

        if(!msg.version().equals(SocksVersion.SOCKS5)) {
            logger.warn(String.format("version wrong : %s , addr :%s", msg.version(),ctx.channel().remoteAddress().toString()));
            ctx.close();
            return;
        }

        Socks5CommandType type = msg.type();
        Socks5AddressType dstAddrType = msg.dstAddrType();
        String addr = msg.dstAddr();
        int port = msg.dstPort();

        if (!type.equals(Socks5CommandType.CONNECT)) {
            ctx.writeAndFlush(new DefaultSocks5CommandResponse(Socks5CommandStatus.COMMAND_UNSUPPORTED,dstAddrType));
            ctx.close();
            return;
        }

        Visitor visitor = new Visitor(ctx.channel());
        if(!visitor.connect(addr,port)){
            ctx.writeAndFlush(new DefaultSocks5CommandResponse(Socks5CommandStatus.FAILURE,dstAddrType));
            ctx.close();
            return;
        }

        ctx.pipeline().addLast(new TransmitHandler(visitor));
        ctx.writeAndFlush(new DefaultSocks5CommandResponse(Socks5CommandStatus.SUCCESS,dstAddrType)).addListener(future -> {
            ctx.pipeline().remove("Socks5CommandRequestDecoder");
            ctx.pipeline().remove(this);
            logger.info(" {} -> socks5Proxy -> {}:{}",ctx.channel().remoteAddress().toString(),addr,port);
        });

    }


    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) throws Exception {
        ctx.close();
        logger.error(cause.getMessage());
    }
}
