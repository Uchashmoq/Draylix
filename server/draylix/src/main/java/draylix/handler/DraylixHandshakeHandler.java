package draylix.handler;

import draylix.cipher.AesCodec;
import draylix.proctocol.Verifier;
import draylix.server.Service;
import draylix.util.AttributeKeyFactory;
import draylix.util.Security;
import io.netty.buffer.ByteBuf;
import io.netty.channel.ChannelFuture;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.util.ReferenceCountUtil;
import io.netty.util.concurrent.Future;
import io.netty.util.concurrent.GenericFutureListener;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;


public class DraylixHandshakeHandler extends SimpleChannelInboundHandler<ByteBuf> {
    public static final byte CONN_PERMIT=0;
    public static final byte CONN_DENIED=1;
    public static final byte NONSUPPORT_SERVICE=2;
    private Verifier verifier;
    private Logger logger= LoggerFactory.getLogger(this.getClass());
    public DraylixHandshakeHandler(Verifier verifier){
        this.verifier=verifier;
    }
    @Override
    protected void channelRead0(ChannelHandlerContext ctx, ByteBuf buf) throws Exception {
        if(buf.readableBytes()<8+8+1){
            ctx.close();
        }
        long timestamp=buf.readLong();
        long announce=buf.readLong();
        byte type = buf.readByte();
        byte[] tokenb = new byte[buf.readableBytes()];
        buf.readBytes(tokenb);
        String token = new String(tokenb);

        if(Security.handShakeCheck==1&&!Security.checkTimeAnno(timestamp,announce)){
            logger.warn("insecure connection : {}",ctx.channel().remoteAddress().toString());
            ctx.close();
            return;
        }

        if(verifier!=null && !verifier.verifyToken(token)){
            reply(ctx,CONN_DENIED,null,null);
            ctx.close();
            logger.info("{} failed to verify,token : {}",ctx.channel().remoteAddress().toString(),token);
            return;
        }

        if(!Service.initService(ctx,type)){
            reply(ctx,NONSUPPORT_SERVICE,null,null);
            ctx.close();
            return;
        }

        byte[] sessionKey=Security.generateRandomBytes(32);
        ctx.channel().attr(AttributeKeyFactory.TOKEN).set(token);

        reply(ctx,CONN_PERMIT,sessionKey,f->{
            ctx.channel().attr(AttributeKeyFactory.DRLX_KEY).set(sessionKey);
            ctx.pipeline().remove(this);
            logger.debug("{} drlx handshake successfully",ctx.channel().remoteAddress().toString());
        } );
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) throws Exception {
        logger.error(cause.getMessage());
    }

    private void reply(ChannelHandlerContext ctx, byte res, byte[] sessionKey, GenericFutureListener<Future<? super Void>> listener){
        ByteBuf data =ctx.alloc().buffer();
        data.writeLong(System.currentTimeMillis()/1000);//时间戳
        data.writeBytes(Security.generateRandomBytes(8));//announce
        data.writeByte(res);
        if (sessionKey!=null) data.writeBytes(sessionKey);
        ChannelFuture future = ctx.writeAndFlush(data);
        if(listener!=null){
            future.addListener(listener);
        }
    }



}
