package draylix.handler;

import io.netty.channel.ChannelHandler;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.codec.socksx.SocksVersion;
import io.netty.handler.codec.socksx.v5.DefaultSocks5InitialRequest;
import io.netty.handler.codec.socksx.v5.DefaultSocks5InitialResponse;
import io.netty.handler.codec.socksx.v5.Socks5AuthMethod;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
@ChannelHandler.Sharable
public class Socks5InitialRequestHandler extends SimpleChannelInboundHandler<DefaultSocks5InitialRequest> {
    private final Logger logger = LoggerFactory.getLogger(this.getClass());
    public static final Socks5InitialRequestHandler INSTANCE = new Socks5InitialRequestHandler();
    @Override
    protected void channelRead0(ChannelHandlerContext ctx, DefaultSocks5InitialRequest msg) throws Exception {
        logger.debug("initial request from {}",ctx.channel().remoteAddress());
        if (msg.decoderResult().isFailure()) {
            logger.error(msg.decoderResult().cause().getMessage());
            ctx.fireChannelRead(msg);
        }
        else{
            if (msg.version().equals(SocksVersion.SOCKS5)){
                ctx.writeAndFlush(new DefaultSocks5InitialResponse(Socks5AuthMethod.NO_AUTH)).addListener(f->{
                    ctx.pipeline().remove("Socks5InitialRequestDecoder");
                    ctx.pipeline().remove(this);
                });
            }else{
                logger.debug("version wrong : {} , addr :{}", msg.version(),ctx.channel().remoteAddress().toString());
            }
        }
    }
}
