import { PropsWithChildren, useEffect, useRef } from 'react';
import { Animated, Easing, StyleSheet } from 'react-native';

type Props = PropsWithChildren<{
  delay?: number;
}>;

export const AnimatedEntry = ({ children, delay = 0 }: Props) => {
  const opacity = useRef(new Animated.Value(0)).current;
  const translateY = useRef(new Animated.Value(20)).current;

  useEffect(() => {
    Animated.parallel([
      Animated.timing(opacity, {
        toValue: 1,
        duration: 350,
        delay,
        easing: Easing.out(Easing.cubic),
        useNativeDriver: true,
      }),
      Animated.timing(translateY, {
        toValue: 0,
        duration: 350,
        delay,
        easing: Easing.out(Easing.cubic),
        useNativeDriver: true,
      }),
    ]).start();
  }, [delay, opacity, translateY]);

  return <Animated.View style={[styles.container, { opacity, transform: [{ translateY }] }]}>{children}</Animated.View>;
};

const styles = StyleSheet.create({
  container: {
    width: '100%',
  },
});
