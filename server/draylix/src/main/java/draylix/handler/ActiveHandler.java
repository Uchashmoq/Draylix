package draylix.handler;

import draylix.util.AttributeKeyFactory;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.ChannelInboundHandlerAdapter;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class ActiveHandler extends ChannelInboundHandlerAdapter {
    private Logger logger = LoggerFactory.getLogger(this.getClass());
    public ActiveHandler(byte[] iv,byte[] ikey){
        this.iv=iv;
        this.ikey=ikey;
    }
    private byte[] iv;
    private byte[] ikey;

    @Override
    public void channelActive(ChannelHandlerContext ctx) throws Exception {
        ctx.channel().attr(AttributeKeyFactory.DRLX_KEY).set(ikey);
        ctx.channel().attr(AttributeKeyFactory.IV).set(iv);
        logger.debug("{} connected",ctx.channel().remoteAddress().toString());
        ctx.pipeline().remove(this);
    }
}
